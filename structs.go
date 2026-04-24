package xtreamcodes

import "encoding/json"

// TODO: Add more flex types on IDs if needed
// for future potential provider issues.

// ServerInfo describes the state of the Xtream-Codes server.
type ServerInfo struct {
	HTTPSPort    FlexInt   `json:"https_port"`
	Port         FlexInt   `json:"port"`
	Process      bool      `json:"process"`
	RTMPPort     FlexInt   `json:"rtmp_port"`
	Protocol     string    `json:"server_protocol"`
	TimeNow      string    `json:"time_now"`
	TimestampNow Timestamp `json:"timestamp_now,string"`
	Timezone     string    `json:"timezone"`
	URL          string    `json:"url"`
	Version      string    `json:"version,omitempty"`
	Revision     string    `json:"revision,omitempty"`
	XUI          bool      `json:"xui,omitempty"`
}

// UserInfo is the current state of the user as it relates to the Xtream-Codes server.
type UserInfo struct {
	ActiveConnections    FlexInt            `json:"active_cons"`
	AllowedOutputFormats []string           `json:"allowed_output_formats"`
	Auth                 ConvertibleBoolean `json:"auth"`
	CreatedAt            Timestamp          `json:"created_at"`
	ExpDate              *Timestamp         `json:"exp_date"`
	IsTrial              ConvertibleBoolean `json:"is_trial,string"`
	MaxConnections       FlexInt            `json:"max_connections"`
	Message              string             `json:"message"`
	Password             string             `json:"password"`
	Status               string             `json:"status"`
	Username             string             `json:"username"`
}

// AuthenticationResponse is a container for what the server returns after the initial authentication.
type AuthenticationResponse struct {
	ServerInfo ServerInfo `json:"server_info"`
	UserInfo   UserInfo   `json:"user_info"`
}

// Category describes a grouping of Stream.
type Category struct {
	ID     FlexInt `json:"category_id"`
	Name   string  `json:"category_name"`
	Parent FlexInt `json:"parent_id"`

	// Set by us, not Xtream.
	Type string `json:"-"`
}

// Stream is a streamble video source.
type Stream struct {
	Added              *Timestamp         `json:"added"`
	CategoryID         FlexInt            `json:"category_id"`
	CategoryIDs        []FlexInt          `json:"category_ids,omitempty"`
	CategoryName       string             `json:"category_name"`
	ContainerExtension string             `json:"container_extension"`
	CustomSid          string             `json:"custom_sid"`
	DirectSource       string             `json:"direct_source,omitempty"`
	EPGChannelID       string             `json:"epg_channel_id"`
	Icon               string             `json:"stream_icon"`
	ID                 FlexInt            `json:"stream_id"`
	IsAdult            ConvertibleBoolean `json:"is_adult"`
	Live               ConvertibleBoolean `json:"live"`
	Name               string             `json:"name"`
	Number             FlexInt            `json:"num"`
	Rating             FlexFloat          `json:"rating"`
	Rating5based       FlexFloat          `json:"rating_5based"`
	SeriesNo           *FlexInt           `json:"series_no,omitempty"`
	TmdbID             FlexInt            `json:"tmdb,omitempty"`
	Trailer            string             `json:"trailer,omitempty"`
	TVArchive          ConvertibleBoolean `json:"tv_archive"`
	TVArchiveDuration  *FlexInt           `json:"tv_archive_duration"`
	Type               string             `json:"stream_type"`
	TypeName           string             `json:"type_name,omitempty"`
}

// SeriesInfo contains information about a TV series.
type SeriesInfo struct {
	BackdropPath   *JSONStringSlice `json:"backdrop_path,omitempty"`
	Cast           string           `json:"cast"`
	CategoryID     *FlexInt         `json:"category_id"`
	CategoryIDs    []FlexInt        `json:"category_ids,omitempty"`
	Cover          string           `json:"cover"`
	Director       string           `json:"director"`
	EpisodeRunTime FlexInt          `json:"episode_run_time"`
	Genre          string           `json:"genre"`
	LastModified   *Timestamp       `json:"last_modified,omitempty"`
	Name           string           `json:"name"`
	Num            FlexInt          `json:"num"`
	Plot           string           `json:"plot"`
	Rating         FlexFloat        `json:"rating"`
	Rating5        FlexFloat        `json:"rating_5based"`
	ReleaseDate    string           `json:"releaseDate"`
	SeriesID       FlexInt          `json:"series_id"`
	StreamType     string           `json:"stream_type"`
	TmdbID         FlexInt          `json:"tmdb,omitempty"`
	YoutubeTrailer string           `json:"youtube_trailer"`
}

// seriesInfoAlias prevents UnmarshalJSON recursion.
type seriesInfoAlias SeriesInfo

// UnmarshalJSON normalises the two release-date key variants providers emit
// (releaseDate and release_date) into a single ReleaseDate field.
func (s *SeriesInfo) UnmarshalJSON(b []byte) error {
	type wire struct {
		*seriesInfoAlias
		ReleaseDateSnake string `json:"release_date"`
	}
	w := wire{seriesInfoAlias: (*seriesInfoAlias)(s)}
	if err := json.Unmarshal(b, &w); err != nil {
		return err
	}
	if s.ReleaseDate == "" {
		s.ReleaseDate = w.ReleaseDateSnake
	}
	return nil
}

// EpisodeInfo contains the metadata block embedded within a SeriesEpisode.
type EpisodeInfo struct {
	AirDate        string           `json:"air_date,omitempty"`
	BackdropPath   *JSONStringSlice `json:"backdrop_path,omitempty"`
	Bitrate        FlexInt          `json:"bitrate,omitempty"`
	Crew           string           `json:"crew,omitempty"`
	DirectedBy     string           `json:"directed_by,omitempty"`
	Duration       string           `json:"duration,omitempty"`
	DurationSecs   FlexInt          `json:"duration_secs,omitempty"`
	ID             FlexInt          `json:"id,omitempty"`
	MovieImage     string           `json:"movie_image"`
	MovieImageTmdb string           `json:"movie_image_tmdb,omitempty"`
	Name           string           `json:"name,omitempty"`
	Overview       string           `json:"overview,omitempty"`
	Plot           string           `json:"plot"`
	Rating         FlexFloat        `json:"rating"`
	ReleaseDate    string           `json:"releasedate"`
	TmdbID         FlexInt          `json:"tmdb_id,omitempty"`
}

type SeriesEpisode struct {
	Added              Timestamp   `json:"added"`
	ContainerExtension string      `json:"container_extension"`
	CustomSid          string      `json:"custom_sid"`
	DirectSource       string      `json:"direct_source"`
	EpisodeNum         FlexInt     `json:"episode_num"`
	ID                 FlexInt     `json:"id"`
	Info               EpisodeInfo `json:"info"`
	Season             FlexInt     `json:"season"`
	Title              string      `json:"title"`
}

// Season describes a single season within a series.
type Season struct {
	AirDate      string  `json:"air_date"`
	Cover        string  `json:"cover"`
	CoverBig     string  `json:"cover_big"`
	CoverTmdb    string  `json:"cover_tmdb"`
	Duration     FlexInt `json:"duration"`
	EpisodeCount FlexInt `json:"episode_count"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	ReleaseDate  string  `json:"releaseDate"`
	SeasonNumber FlexInt `json:"season_number"`
}

// seasonAlias prevents UnmarshalJSON recursion.
type seasonAlias Season

// UnmarshalJSON normalises the two release-date key variants into ReleaseDate.
func (s *Season) UnmarshalJSON(b []byte) error {
	type wire struct {
		*seasonAlias
		ReleaseDateSnake string `json:"release_date"`
	}
	w := wire{seasonAlias: (*seasonAlias)(s)}
	if err := json.Unmarshal(b, &w); err != nil {
		return err
	}
	if s.ReleaseDate == "" {
		s.ReleaseDate = w.ReleaseDateSnake
	}
	return nil
}

type Series struct {
	Episodes map[string][]SeriesEpisode `json:"episodes"`
	Info     SeriesInfo                 `json:"info"`
	Seasons  []Season                   `json:"seasons"`
}

// VODInfo is the metadata block within a VideoOnDemandInfo response.
type VODInfo struct {
	Actors         string    `json:"actors"`
	Age            string    `json:"age"`
	BackdropPath   []string  `json:"backdrop_path"`
	Bitrate        FlexInt   `json:"bitrate"`
	Cast           string    `json:"cast"`
	Country        string    `json:"country"`
	Cover          string    `json:"cover,omitempty"`
	CoverBig       string    `json:"cover_big"`
	Description    string    `json:"description"`
	Director       string    `json:"director"`
	Duration       string    `json:"duration"`
	DurationSecs   FlexInt   `json:"duration_secs"`
	EpisodeRunTime *FlexInt  `json:"episode_run_time,omitempty"`
	Genre          string    `json:"genre"`
	KinopoiskURL   string    `json:"kinopoisk_url,omitempty"`
	MovieImage     string    `json:"movie_image"`
	Name           string    `json:"name"`
	OriginalName   string    `json:"o_name"`
	Plot           string    `json:"plot"`
	Rating         FlexFloat `json:"rating"`
	ReleaseDate    string    `json:"releasedate"`
	Runtime        string    `json:"runtime,omitempty"`
	Status         string    `json:"status"`
	TmdbID         FlexInt   `json:"tmdb_id"`
	Year           FlexInt   `json:"year,omitempty"`
	YoutubeTrailer string    `json:"youtube_trailer"`
}

// VODMovieData is the stream identity block within a VideoOnDemandInfo response.
type VODMovieData struct {
	Added              Timestamp `json:"added"`
	CategoryID         FlexInt   `json:"category_id"`
	CategoryIDs        []FlexInt `json:"category_ids,omitempty"`
	ContainerExtension string    `json:"container_extension"`
	CustomSid          string    `json:"custom_sid"`
	DirectSource       string    `json:"direct_source"`
	Name               string    `json:"name"`
	StreamID           FlexInt   `json:"stream_id"`
}

// VideoOnDemandInfo contains information about a video on demand stream.
type VideoOnDemandInfo struct {
	Info      VODInfo      `json:"info"`
	MovieData VODMovieData `json:"movie_data"`
}

// UnmarshalJSON tolerates providers that emit "info": [] (empty array) when
// the metadata block is absent for a given record — treat as a zero-value
// VODInfo rather than failing the whole decode.
func (v *VideoOnDemandInfo) UnmarshalJSON(b []byte) error {
	type wire struct {
		Info      json.RawMessage `json:"info"`
		MovieData json.RawMessage `json:"movie_data"`
	}
	var w wire
	if err := json.Unmarshal(b, &w); err != nil {
		return err
	}
	if len(w.Info) > 0 {
		if err := json.Unmarshal(w.Info, &v.Info); err != nil {
			var placeholder []struct{}
			if arrErr := json.Unmarshal(w.Info, &placeholder); arrErr != nil {
				return err
			}
		}
	}
	if len(w.MovieData) > 0 {
		if err := json.Unmarshal(w.MovieData, &v.MovieData); err != nil {
			return err
		}
	}
	return nil
}

type epgContainer struct {
	EPGListings []EPGInfo `json:"epg_listings"`
}

// EPGInfo describes electronic programming guide information of a stream.
type EPGInfo struct {
	ChannelID      string             `json:"channel_id"`
	Description    Base64Value        `json:"description"`
	End            string             `json:"end"`
	EPGID          FlexInt            `json:"epg_id"`
	HasArchive     ConvertibleBoolean `json:"has_archive"`
	ID             FlexInt            `json:"id"`
	Lang           string             `json:"lang"`
	NowPlaying     ConvertibleBoolean `json:"now_playing"`
	Start          string             `json:"start"`
	StartTimestamp Timestamp          `json:"start_timestamp"`
	StopTimestamp  Timestamp          `json:"stop_timestamp"`
	Title          Base64Value        `json:"title"`
}
