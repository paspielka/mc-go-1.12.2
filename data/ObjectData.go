package data

const (
	Item             = 2
	Minecarts        = 10
	Arrow            = 60
	Fireball         = 63
	SmallFireball    = 64
	WitherSkull      = 66
	ShulkerBullet    = 67
	LlamaSpit        = 68
	FallingBlock     = 70
	ItemFrame        = 71
	Potion           = 73
	ExperienceBottle = 75
	FishingFloat     = 90
	SpectralArrow    = 91
	DragonFireball   = 93
)

type CreateObject struct {
	EntityID int32
	ObjectID [2]int64
	TypeID   byte
	X        float64
	Y        float64
	Z        float64
	Pitch    float64
	Yaw      float64
	Data     int8
	VelX     float64
	VelY     float64
	VelZ     float64
}
