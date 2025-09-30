package models

import "time"

type UsersStats struct {
	UserStatsID              string    `gorm:"primaryKey;type:varchar(25);column:user_stats_id" json:"userStatsId"`
	UserID                   string    `gorm:"type:varchar(25);not null;unique;column:user_id" json:"userId"`
	FollowersCount           int       `gorm:"default:0;column:followers_count" json:"followersCount"`
	FollowingsCount          int       `gorm:"default:0;column:followings_count" json:"followingsCount"`
	UnReadNotificationsCount int       `gorm:"default:0;column:un_read_notifications_count" json:"unReadNotificationsCount"`
	CartItemsCount           int       `gorm:"default:0;column:cart_items_count" json:"cartItemsCount"`
	OrdersCount              int       `gorm:"default:0;column:orders_count" json:"ordersCount"`
	PendingOrdersCount       int       `gorm:"default:0;column:pending_orders_count" json:"pendingOrdersCount"`
	ApprovedOrdersCount      int       `gorm:"default:0;column:approved_orders_count" json:"approvedOrdersCount"`
	DeliveredOrdersCount     int       `gorm:"default:0;column:delivered_orders_count" json:"deliveredOrdersCount"`
	CanceledOrdersCount      int       `gorm:"default:0;column:canceled_orders_count" json:"canceledOrdersCount"`
	PreOrdersCount           int       `gorm:"default:0;column:pre_orders_count" json:"preOrdersCount"`
	RequestedPreOrdersCount  int       `gorm:"default:0;column:requested_pre_orders_count" json:"requestedPreOrdersCount"`
	RefunededPreOrdersCount  int       `gorm:"default:0;column:refunded_pre_orders_count" json:"refundedPreOrdersCount"`
	RefunededOrdersCount     int       `gorm:"default:0;column:refunded_orders_count" json:"refundedOrdersCount"`
	ApprovedPreOrdersCount   int       `gorm:"default:0;column:approved_pre_orders_count" json:"approvedPreOrdersCount"`
	DeclinedPreOrdersCount   int       `gorm:"default:0;column:declined_pre_orders_count" json:"declinedPreOrdersCount"`
	PendingPreOrdersCount    int       `gorm:"default:0;column:pending_pre_orders_count" json:"pendingPreOrdersCount"`
	DeliveredPreOrdersCount  int       `gorm:"default:0;column:delivered_pre_orders_count" json:"deliveredPreOrdersCount"`
	CanceledPreOrdersCount   int       `gorm:"default:0;column:canceled_pre_orders_count" json:"canceledPreOrdersCount"`
	TotalLikes               int       `gorm:"default:0;column:total_likes" json:"totalLikes"`
	TotalPosts               int       `gorm:"default:0;column:total_posts" json:"totalPosts"`
	CreatedAt                time.Time `gorm:"not null;column:created_at" json:"createdAt"`
	UpdatedAt                time.Time `gorm:"column:updated_at" json:"updatedAt"`
	CreatedBy                string    `gorm:"type:varchar(80);not null;column:created_by" json:"createdBy"`
	UpdatedBy                string    `gorm:"type:varchar(80);not null;column:updated_by" json:"updatedBy"`

	User *User `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`
}

func (UsersStats) TableName() string {
	return "users_stats"
}
