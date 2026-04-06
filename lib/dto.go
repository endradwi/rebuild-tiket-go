package lib

import "time"


type User struct {
	Id int 
	Email string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
	RoleId int `json:"role_id"`
	RoleName string `json:"role_name,omitempty"`
}

type UserRole struct {
	User User
	RoleId int `json:"role_id" form:"role_id"`
}

type RegisterRequest struct {
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required"`
}

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  any    `json:"result,omitempty"`
}

type ListResponse struct {
	Status   int       `json:"status"`
	Message  string    `json:"message"`
	Result   any       `json:"result,omitempty"`
	PageInfo *PageInfo `json:"page_info,omitempty"`
}

type PageInfo struct {
	CurentPage int `json:"current_page"`
	NextPage   int `json:"next_page"`
	PrevPage   int `json:"prev_page"`
	TotalPage  int `json:"total_page"`
	TotalData  int `json:"total_data"`
}

type ResetPassword struct {
	Id int
	ProfileId int
	TokenHash string
	ExpiredAt time.Time
	UsedAt *time.Time
	CreatedAt time.Time
}

type ResetPasswordRequest struct {
	Token string `json:"token" form:"token"`
	Password string `json:"password" form:"password"`
}

type UserProfile struct {
	Id int `json:"id"`
	Email string `json:"email"`
	FirstName *string `json:"first_name"`
	LastName *string `json:"last_name"`
	PhoneNumber *string `json:"phone_number"`
	Image *string `json:"image"`
	Point *int `json:"point"`
}

type ProfileUpdateRequest struct {
	FirstName *string `form:"first_name"`
	LastName *string `form:"last_name"`
	PhoneNumber *string `form:"phone_number"`
	Image *string `form:"-"`
}

type Movie struct {
	Id             int        `json:"id"`
	Image          *string    `json:"image"`
	Title          string     `json:"title"`
	ReleasedAt     *time.Time `json:"released_at"`
	Recommendation *bool      `json:"recommendation"`
	Duration       *string    `json:"duration"`
	Synopsis       *string    `json:"synopsis"`
	DirectorName   *string    `json:"director_name"`
	GenreId        *int       `json:"genre_id"`
	GenreName      string     `json:"genre_name"`
	CasterId       *int       `json:"caster_id"`
	CasterName     string     `json:"caster_name"`
	Genres         []string            `json:"genres,omitempty"`
	Casters        []string            `json:"casters,omitempty"`
	Cinemas        []MovieCinemaDetail `json:"cinemas,omitempty"`
}

type MovieCinemaDetail struct {
	CinemaId     int              `json:"cinema_id"`
	CinemaName   string           `json:"cinema_name"`
	CinemaImage  *string          `json:"cinema_image"`
	LocationName string           `json:"location_name"`
	Showtimes    []CinemaShowtime `json:"showtimes"`
}

type CinemaShowtime struct {
	ShowtimeId int    `json:"showtime_id"`
	ShowDate   string `json:"show_date"`
	ShowTime   string `json:"show_time"`
	Price      int    `json:"price"`
}

type ShowtimeReq struct {
	Date  time.Time `json:"date" form:"date"`
	Times []string  `json:"times" form:"times"`
	Price int       `json:"price" form:"price"`
}

type MovieCreateRequest struct {
	Title          string     `form:"title" binding:"required"`
	ReleasedAt     *time.Time `form:"released_at"`
	Recommendation *bool      `form:"recommendation"`
	Duration       *string    `form:"duration"`
	Synopsis       *string    `form:"synopsis"`
	DirectorName   *string    `form:"director_name"`
	GenreIds       []int      `form:"genre_ids"`
	CasterIds      []int      `form:"caster_ids"`
	CinemaIds      []int      `form:"cinema_ids"`
	Showtimes      []ShowtimeReq `json:"showtimes" form:"showtimes"`
	Image          *string    `form:"-"`
}

type MovieUpdateRequest struct {
	Title          *string    `form:"title"`
	ReleasedAt     *time.Time `form:"released_at"`
	Recommendation *bool      `form:"recommendation"`
	Duration       *string    `form:"duration"`
	Synopsis       *string    `form:"synopsis"`
	DirectorName   *string    `form:"director_name"`
	GenreIds       []int      `form:"genre_ids"`
	CasterIds      []int      `form:"caster_ids"`
	CinemaIds      []int      `form:"cinema_ids"`
	Image          *string    `form:"-"`
}

type MovieQueryParams struct {
	Limit  int    `form:"limit"`
	Page   int    `form:"page"`
	Search string `form:"search"`
	Sort   string `form:"sort"` // "asc" or "desc", default "asc"
	Month  int    `form:"month"` // For admin filtering
	Year   int    `form:"year"`
}

type SalesStat struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type DashboardStats struct {
	SalesChart      []SalesStat `json:"sales_chart"`
	TicketSales     []SalesStat `json:"ticket_sales"` // By Category/Location
	AverageEarnings int         `json:"average_earnings"`
}

type MovieDetailParams struct {
	LocationId int    `form:"location_id"`
	Date       string `form:"date"`
	Time       string `form:"time"`
}

type Location struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Cinema struct {
	Id           int        `json:"id"`
	CinemaName   string     `json:"cinema_name"`
	Image        *string    `json:"image"`
	LocationId   *int       `json:"location_id"`
	LocationName string     `json:"location_name,omitempty"`
}

type MovieShowtime struct {
	Id         int       `json:"id"`
	MovieId    int       `json:"movie_id"`
	MovieTitle string    `json:"movie_title"`
	CinemaId   int       `json:"cinema_id"`
	CinemaName string    `json:"cinema_name"`
	LocationName string  `json:"location_name,omitempty"`
	ShowDate   time.Time `json:"show_date"`
	ShowTime   string    `json:"show_time"`
	Price      int       `json:"price"`
}

type Seat struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type SeatWithStatus struct {
	Seat
	IsOccupied bool `json:"is_occupied"`
}

type Order struct {
	Id          int           `json:"id"`
	OrderNumber string        `json:"order_number"`
	ProfileId   *int          `json:"profile_id"`
	ShowtimeId  int           `json:"showtime_id"`
	FullName    *string       `json:"full_name"`
	Email       *string       `json:"email"`
	PhoneNumber *string       `json:"phone_number"`
	TotalPrice  int           `json:"total_price"`
	Status      string        `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	Seats       []Seat        `json:"seats,omitempty"`
	MovieTitle  string        `json:"movie_title,omitempty"`
	CinemaName   string        `json:"cinema_name,omitempty"`
	CinemaImage  *string       `json:"cinema_image,omitempty"`
	LocationName string        `json:"location_name,omitempty"`
	ShowDate    *time.Time    `json:"show_date,omitempty"`
	ShowTime    *string       `json:"show_time,omitempty"`
	InvoiceUrl  *string       `json:"invoice_url,omitempty"`
	ExternalId  *string       `json:"external_id,omitempty"`
}

type OrderCreateRequest struct {
	ShowtimeId int   `json:"showtime_id" binding:"required"`
	SeatIds    []int `json:"seat_ids" binding:"required"`
}

type PaymentRequest struct {
	FullName      string `json:"full_name" binding:"required"`
	Email         string `json:"email" binding:"required"`
	PhoneNumber   string `json:"phone_number" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

type Payment struct {
	Id             int       `json:"id"`
	OrderId        int       `json:"order_id"`
	TotalPayment   int       `json:"total_payment"`
	PaymentMethod  string    `json:"payment_method"`
	PaymentStatus  string    `json:"payment_status"`
	ExpiredAt      time.Time `json:"expired_at"`
	QrCode         *string   `json:"qr_code"`
	InvoiceUrl     *string   `json:"invoice_url"`
	ExternalId     *string   `json:"external_id"`
}

type XenditWebhookRequest struct {
	Id                 string    `json:"id"`
	ExternalId         string    `json:"external_id"`
	Status             string    `json:"status"`
	Amount             float64   `json:"amount"`
	PayerEmail         string    `json:"payer_email"`
	Description        string    `json:"description"`
	PaymentMethod      string    `json:"payment_method"`
	PaymentChannel     string    `json:"payment_channel"`
	PaidAt            *time.Time `json:"paid_at"`
	Created           *time.Time `json:"created"`
	Updated           *time.Time `json:"updated"`
}

type Genre struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Caster struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}


