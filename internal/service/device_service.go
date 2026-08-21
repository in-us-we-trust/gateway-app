package service

import (
	"context"
	"gateway-app/internal/client"
	"gateway-app/internal/domain"
)

type DeviceService struct {
	userClient      client.UserClient
	detectiveClient client.DetectorClient
}

func NewDeviceService(
	userClient client.UserClient,
	detectiveClient client.DetectorClient,
) DeviceService {
	return DeviceService{
		userClient:      userClient,
		detectiveClient: detectiveClient,
	}
}

func (s *DeviceService) GetMyDevice(
	ctx context.Context,
	userID string,
) (*domain.DeviceResponseDto, error) {

	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	detectiveDevice, err := s.detectiveClient.GetDevice(ctx, device.DeviceID)
	if err != nil {
		return nil, err
	}

	return domain.ToDeviceResponseDto(device, detectiveDevice), nil
}

func (s *DeviceService) LinkDevice(
	ctx context.Context,
	userID string,
	req domain.LinkDeviceRequest,
) (*domain.DeviceResponseDto, error) {

	device, err := s.userClient.LinkDevice(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	detectiveDevice, err := s.detectiveClient.ClaimDevice(ctx, device.DeviceID)
	if err != nil {
		return nil, err
	}

	return domain.ToDeviceResponseDto(device, detectiveDevice), nil
}

func (s *DeviceService) UnlinkDevice(
	ctx context.Context,
	userID string,
) error {

	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return err
	}

	if err := s.userClient.UnlinkDevice(ctx, userID); err != nil {
		return err
	}

	return s.detectiveClient.UnclaimDevice(ctx, device.DeviceID)
}

func (s *DeviceService) UpdateMyDevice(
	ctx context.Context,
	userID string,
	req domain.UpdateDeviceRequest,
) (*domain.DeviceResponseDto, error) {

	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	detectiveDevice, err := s.detectiveClient.UpdateDevice(ctx, device.DeviceID, req)
	if err != nil {
		return nil, err
	}

	return domain.ToDeviceResponseDto(device, detectiveDevice), nil
}

func (s *DeviceService) GetMyDeviceEvents(
	ctx context.Context,
	userID string,
) (*domain.TracksResult, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	tracks, err := s.detectiveClient.GetDeviceTracks(ctx, device.DeviceID)
	if err != nil {
		return nil, err
	}

	return splitUrgent(*tracks), nil
}

func (s *DeviceService) GetMyDeviceEventDetail(
	ctx context.Context,
	userID string,
	trackID string,
) (*domain.TrackDetail, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.detectiveClient.GetTrackDetail(ctx, device.DeviceID, trackID)
}

// severityRank ordena las severities. Mayor número = más urgente.
var severityRank = map[string]int{
	"NORMAL":     0,
	"LLAMATIVO":  1,
	"SOSPECHOSO": 2,
	"PELIGROSO":  3,
}

// splitUrgent separa el track más urgente (mayor severity; ante empate, el más
// reciente por createdAt) del resto. El urgente NO queda en la lista de events.
func splitUrgent(tracks []domain.Track) *domain.TracksResult {
	if len(tracks) == 0 {
		return &domain.TracksResult{UrgentEvent: nil, Events: []domain.Track{}}
	}

	urgentIdx := 0
	for i := 1; i < len(tracks); i++ {
		if isMoreUrgent(tracks[i], tracks[urgentIdx]) {
			urgentIdx = i
		}
	}

	urgent := tracks[urgentIdx]
	rest := make([]domain.Track, 0, len(tracks)-1)
	for i, t := range tracks {
		if i != urgentIdx {
			rest = append(rest, t)
		}
	}

	return &domain.TracksResult{UrgentEvent: &urgent, Events: rest}
}

// isMoreUrgent: mayor severity gana; ante empate, el createdAt más reciente.
func isMoreUrgent(candidate, current domain.Track) bool {
	cRank := severityRank[candidate.FinalSeverity]
	curRank := severityRank[current.FinalSeverity]
	if cRank != curRank {
		return cRank > curRank
	}
	return candidate.CreatedAt > current.CreatedAt
}

func (s *DeviceService) RegisterPushToken(
	ctx context.Context,
	userID string,
	req domain.RegisterPushTokenRequest,
) (*domain.AppDeviceResponse, error) {
	return s.userClient.RegisterPushToken(ctx, userID, req)
}
func (s *DeviceService) DeletePushToken(
	ctx context.Context,
	userID string,
	appDeviceID int64,
) error {
	return s.userClient.DeletePushToken(ctx, userID, appDeviceID)
}
