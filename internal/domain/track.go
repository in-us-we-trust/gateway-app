package domain

// Track es la visita analizada que devuelve el detective-manager en /tracks.
// El gateway hace de proxy: expone estos campos tal cual, sin transformar.
type Track struct {
	TrackID       string `json:"trackId"`
	Code          string `json:"code"`
	Title         string `json:"title"`
	FinalVerdict  string `json:"finalVerdict"`
	FinalSeverity string `json:"finalSeverity"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// TracksResult separa el evento más urgente del resto de la lista.
// UrgentEvent es el track de mayor severity (y, ante empate, el más reciente);
// puede ser nil si no hay tracks. Events son los demás (sin el urgente).
type TracksResult struct {
	UrgentEvent *Track  `json:"urgentEvent"`
	Events      []Track `json:"events"`
}

// TrackDetail es el detalle de una visita que devuelve el detective-manager
// en /tracks/{trackId}. El gateway hace de proxy: expone estos campos tal cual.
type TrackDetail struct {
	Code          string  `json:"code"`
	Title         string  `json:"title"`
	FinalVerdict  string  `json:"finalVerdict"`
	FinalSeverity string  `json:"finalSeverity"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
	ImageURL      *string `json:"imageUrl"`
}
