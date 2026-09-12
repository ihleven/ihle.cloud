package hi

// The HiDrive API's own object model: what /meta and /dir return.

type Meta struct {
	ID             string `json:"id,omitempty"`
	NameURLEncoded string `json:"name"`
	Path           string `json:"path,omitempty"`
	Type_          string `json:"type,omitempty"`
	Size_          int    `json:"size,omitempty"`
	Category       string `json:"category,omitempty"`
	NMembers       int    `json:"nmembers,omitempty"`
	MTime          int64  `json:"mtime,omitempty"`
	Members        []Meta `json:"members,omitempty"`
	Mimetype       string `json:"mime_type,omitempty"`

	CTime    int    `json:"ctime,omitempty"`
	Readable bool   `json:"readable,omitempty"`
	Writable bool   `json:"writable,omitempty"`
	ParentID string `json:"parent_id,omitempty"`

	Image *Image `json:"image,omitempty"`
}

type Image struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Exif   *Exif `json:"exif"`
}

type Exif struct {
	DateTimeOriginal string  `json:",omitempty"`
	Make             string  `json:",omitempty"`
	Model            string  `json:",omitempty"`
	ImageWidth       int     `json:",omitempty"`
	ImageHeight      int     `json:",omitempty"`
	ExifImageWidth   int     `json:",omitempty"`
	ExifImageHeight  int     `json:",omitempty"`
	Aperture         float64 `json:",omitempty"`
	ExposureTime     float64 `json:",omitempty"`
	ISO              int     `json:",omitempty"`
	FocalLength      float64 `json:",omitempty"`
	Orientation      int     `json:",omitempty"`
	XResolution      float64 `json:",omitempty"`
	YResolution      float64 `json:",omitempty"`
	ResolutionUnit   int     `json:",omitempty"`
	BitsPerSample    int     `json:",omitempty"`
	GPSLatitude      float64 `json:",omitempty"`
	GPSLongitude     float64 `json:",omitempty"`
	GPSAltitude      float64 `json:",omitempty"`
}
