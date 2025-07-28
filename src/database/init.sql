-- Create Role
CREATE ROLE phinex LOGIN PASSWORD 'phinex';
ALTER ROLE phinex SET TIME ZONE 'UTC';

-- Create schemas
CREATE SCHEMA IF NOT EXISTS public;

-- Create ENUM types in public schema
CREATE TYPE public.user_status AS ENUM ('active', 'banned', 'suspended', 'online');
CREATE TYPE public.role_name AS ENUM ('Authenticated', 'Anonymous', 'BusinessOwner', 'SuperAdmin', 'PaymentAgent', 'Admin');
CREATE TYPE public.business_status AS ENUM ('active', 'banned', 'suspended', 'online', 'requested');
CREATE TYPE public.notification_type AS ENUM ('business', 'user', 'payment', 'transaction', 'other');
CREATE TYPE public.order_status AS ENUM ('pending', 'canceled', 'delivered', 'approved');
CREATE TYPE public.payment_method AS ENUM ('completed', 'refunded', 'pending');
CREATE TYPE public.preorder_status AS ENUM ('requested', 'pending', 'approved', 'declined', 'canceled', 'delivered');
CREATE TYPE public.product_type AS ENUM ('order', 'preorder');
CREATE TYPE public.transaction_type AS ENUM ('deposit', 'withdrawal', 'transfer', 'payment', 'refund');
CREATE TYPE public.transaction_status AS ENUM ('pending', 'completed', 'failed', 'canceled');
CREATE TYPE public.refund_request_status AS ENUM ('pending', 'approved', 'rejected', 'processed');
CREATE TYPE public.subscription_status AS ENUM ('active', 'cancelled', 'expired');
CREATE TYPE public.subscription_plan AS ENUM ('monthly', 'yearly');

-- Create tables in dependency order

CREATE TABLE IF NOT EXISTS public.users (
    user_id VARCHAR(25) PRIMARY KEY,
    user_id_serial SERIAL UNIQUE,
    first_name VARCHAR NOT NULL,
    middle_name VARCHAR,
    full_name VARCHAR,
    user_name VARCHAR NOT NULL,
    last_name VARCHAR NOT NULL,
    profile_image TEXT,
    bio TEXT,
    phone_number VARCHAR,
    status public.user_status DEFAULT 'active',
    password VARCHAR NOT NULL,
    gender VARCHAR NOT NULL,
    dob DATE NOT NULL,
    email VARCHAR NOT NULL UNIQUE,
    verified BOOLEAN DEFAULT FALSE,
    email_is_verified BOOLEAN DEFAULT FALSE,
    phone_number_is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL
);

CREATE TABLE IF NOT EXISTS public.roles (
    role_id VARCHAR(25) PRIMARY KEY,
    role_id_serial SERIAL UNIQUE,
    role_name public.role_name NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL
);

CREATE TABLE IF NOT EXISTS public.business (
    business_id VARCHAR(25) PRIMARY KEY,
    business_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25) NOT NULL UNIQUE,
    name VARCHAR,
    logo TEXT,
    description TEXT,
    niche VARCHAR,
    status public.business_status DEFAULT 'requested',
    phone_numbers JSON,
    email VARCHAR NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.blogs (
    blog_id VARCHAR(25) PRIMARY KEY,
    blog_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25) NOT NULL,
    slug VARCHAR(25),
    title VARCHAR,
    url TEXT,
    external_link TEXT,
    external_link_title VARCHAR,
    likes_count INTEGER NOT NULL DEFAULT 0,
    comments_count INTEGER NOT NULL DEFAULT 0,
    shares_count INTEGER NOT NULL DEFAULT 0,
    is_reel BOOLEAN DEFAULT FALSE,
    views_count INTEGER DEFAULT 0,
    text TEXT,
    images JSON,
    video TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.wallets (
    wallet_id VARCHAR(25) PRIMARY KEY,
    wallet_id_serial SERIAL UNIQUE NOT NULL,
    account_number VARCHAR(20),
    user_id VARCHAR(25) NOT NULL UNIQUE,
    balance NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'SLE',
    is_active BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(80) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(80) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.products (
    product_id VARCHAR(25) PRIMARY KEY,
    product_id_serial SERIAL UNIQUE,
    business_id VARCHAR(25) NOT NULL,
    location VARCHAR,
    category VARCHAR DEFAULT 'Others',
    type public.product_type,
    variants JSON,
    sizes JSON,
    available BOOLEAN DEFAULT TRUE,
    price NUMERIC(15,2) NOT NULL,
    initial_price NUMERIC(15,2),
    rating FLOAT DEFAULT 0.0,
    url TEXT,
    name VARCHAR,
    images JSON,
    video TEXT,
    description TEXT,
    quantity INTEGER NOT NULL DEFAULT 1,
    number_of_days INTEGER,
    pre_order_date TIMESTAMPTZ,
    views_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (business_id) REFERENCES public.business(business_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.user_roles (
    user_role_id VARCHAR(25) PRIMARY KEY,
    user_role_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25) NOT NULL,
    role_id VARCHAR(25) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (role_id) REFERENCES public.roles(role_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.user_auths (
    user_auth_id VARCHAR(25) PRIMARY KEY,
    user_auth_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25) NOT NULL UNIQUE,
    auth_provider VARCHAR(80),
    otp TEXT,
    otp_expiry TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    last_login_ip TEXT,
    created_by VARCHAR(80) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(80) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.user_settings (
    user_setting_id VARCHAR(25) PRIMARY KEY,
    user_setting_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25) NOT NULL UNIQUE,
    push_notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    email_notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    theme VARCHAR(20) NOT NULL DEFAULT 'light',
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    profile_visibility VARCHAR(20) NOT NULL DEFAULT 'public',
    location_sharing_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    auto_refresh_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    settings JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.encryption_keys (
    encryption_key_id VARCHAR(25) PRIMARY KEY,
    encryption_key_id_serial SERIAL UNIQUE,
    public_key TEXT,
    private_key TEXT,
    user_id VARCHAR(25) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.users_stats (
    user_stats_id VARCHAR(25) PRIMARY KEY,
    user_stats_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25) NOT NULL UNIQUE,
    followers_count INTEGER DEFAULT 0,
    followings_count INTEGER DEFAULT 0,
    un_read_notifications_count INTEGER DEFAULT 0,
    cart_items_count INTEGER DEFAULT 0,
    orders_count INTEGER DEFAULT 0,
    pending_orders_count INTEGER DEFAULT 0,
    approved_orders_count INTEGER DEFAULT 0,
    delivered_orders_count INTEGER DEFAULT 0,
    canceled_orders_count INTEGER DEFAULT 0,
    pre_orders_count INTEGER DEFAULT 0,
    requested_pre_orders_count INTEGER DEFAULT 0,
    approved_pre_orders_count INTEGER DEFAULT 0,
    declined_pre_orders_count INTEGER DEFAULT 0,
    pending_pre_orders_count INTEGER DEFAULT 0,
    delivered_pre_orders_count INTEGER DEFAULT 0,
    canceled_pre_orders_count INTEGER DEFAULT 0,
    total_likes INTEGER DEFAULT 0,
    total_posts INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.business_locations (
    location_id VARCHAR(25) PRIMARY KEY,
    location_id_serial SERIAL UNIQUE,
    business_id VARCHAR(25) NOT NULL,
    address VARCHAR NOT NULL,
    city VARCHAR NOT NULL,
    province VARCHAR NOT NULL,
    latitude NUMERIC(10,8),
    longitude NUMERIC(11,8),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (business_id) REFERENCES public.business(business_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.business_stats (
    business_stats_id VARCHAR(25) PRIMARY KEY,
    business_stats_id_serial SERIAL UNIQUE,
    business_id VARCHAR(25) NOT NULL UNIQUE,
    products_count INTEGER DEFAULT 0,
    order_product_type_count INTEGER DEFAULT 0,
    pre_order_product_type_count INTEGER DEFAULT 0,
    orders_count INTEGER DEFAULT 0,
    approved_orders_count INTEGER DEFAULT 0,
    delivered_orders_count INTEGER DEFAULT 0,
    pre_orders_count INTEGER DEFAULT 0,
    requested_pre_orders_count INTEGER DEFAULT 0,
    approved_pre_orders_count INTEGER DEFAULT 0,
    declined_pre_orders_count INTEGER DEFAULT 0,
    total_sales NUMERIC(15,2) DEFAULT 0.00,
    total_revenue NUMERIC(15,2) DEFAULT 0.00,
    active_products_count INTEGER DEFAULT 0,
    inactive_products_count INTEGER DEFAULT 0,
    pending_orders_count INTEGER DEFAULT 0,
    canceled_orders_count INTEGER DEFAULT 0,
    pending_pre_orders_count INTEGER DEFAULT 0,
    delivered_pre_orders_count INTEGER DEFAULT 0,
    canceled_pre_orders_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (business_id) REFERENCES public.business(business_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.orders (
    order_id VARCHAR(25) PRIMARY KEY,
    order_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25),
    business_id VARCHAR(25) NOT NULL,
    business_name VARCHAR,
    business_logo VARCHAR,
    delivery_id VARCHAR,
    delivery_name VARCHAR,
    delivery_address VARCHAR,
    delivery_phone_number VARCHAR,
    delivery_latitude NUMERIC(10,8),
    delivery_longitude NUMERIC(11,8),
    payment_method VARCHAR,
    payment_id VARCHAR,
    deleted_by_user BOOLEAN DEFAULT FALSE,
    deleted_by_business BOOLEAN DEFAULT FALSE,
    opened_by_user BOOLEAN DEFAULT FALSE,
    opened_by_business BOOLEAN DEFAULT FALSE,
    status public.order_status DEFAULT 'pending',
    payment_status public.payment_method DEFAULT 'pending',
    total_price NUMERIC(15,2) NOT NULL DEFAULT '0.00',
  quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (business_id) REFERENCES public.business(business_id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS public.order_items (
    order_item_id VARCHAR(25) PRIMARY KEY,
    order_item_id_serial SERIAL UNIQUE,
    order_id VARCHAR(25) NOT NULL,
    product_id VARCHAR(25) NOT NULL,
    product_name VARCHAR,
    product_image TEXT,
    product_price NUMERIC(15,2) NOT NULL,
  quantity INTEGER NOT NULL DEFAULT 1,
     total_price NUMERIC(15,2) NOT NULL DEFAULT '0.00',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES public.orders(order_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (product_id) REFERENCES public.products(product_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.carts (
    cart_id VARCHAR(25) PRIMARY KEY,
    cart_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25),
    business_id VARCHAR(25) NOT NULL,
    total_price NUMERIC(15,2) NOT NULL DEFAULT '0.00',
    quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (business_id) REFERENCES public.business(business_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.cart_items (
    cart_item_id VARCHAR(25) PRIMARY KEY,
    cart_item_id_serial SERIAL UNIQUE,
    cart_id VARCHAR(25) NOT NULL,
    product_id VARCHAR(25) NOT NULL,
    product_name VARCHAR,
    product_image TEXT,
    product_price NUMERIC(15,2) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    total_price NUMERIC(15,2) NOT NULL DEFAULT '0.00',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (cart_id) REFERENCES public.carts(cart_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (product_id) REFERENCES public.products(product_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.comments (
    comment_id VARCHAR(25) PRIMARY KEY,
    comment_id_serial SERIAL UNIQUE,
    ref_id VARCHAR(25),
    user_id VARCHAR(25),
    text TEXT,
    image TEXT,
    likes_count INTEGER DEFAULT 0,
    replies_count INTEGER DEFAULT 0,
    sticker TEXT,
    video TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.views (
    view_id VARCHAR(25) PRIMARY KEY,
    view_id_serial SERIAL UNIQUE,
    ref_id VARCHAR(25) NOT NULL,
    user_id VARCHAR(25),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.pinned_blogs (
    pinned_blog_id VARCHAR(25) PRIMARY KEY,
    pinned_blog_id_serial SERIAL UNIQUE,
    blog_id VARCHAR(25) NOT NULL UNIQUE,
    user_id VARCHAR(25) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (blog_id) REFERENCES public.blogs(blog_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.pinned_products (
    pinned_product_id VARCHAR(25) PRIMARY KEY,
    pinned_product_id_serial SERIAL UNIQUE,
    product_id VARCHAR(25) NOT NULL UNIQUE,
    user_id VARCHAR(25) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (product_id) REFERENCES public.products(product_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.contacts (
    contact_id VARCHAR(25) PRIMARY KEY,
    contact_id_serial SERIAL UNIQUE,
    country VARCHAR,
    city VARCHAR,
    locations JSON,
    latitude DECIMAL,
    longitude DECIMAL,
    permanent_address VARCHAR,
    current_address VARCHAR,
    phone_numbers JSON,
    business_id VARCHAR(25) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (business_id) REFERENCES public.business(business_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.follows (
    follow_id VARCHAR(25) PRIMARY KEY,
    follow_id_serial SERIAL UNIQUE,
    follower_id VARCHAR(25) NOT NULL,
    following_id VARCHAR(25) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (follower_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (following_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.notification_details (
    notification_detail_id VARCHAR(25) PRIMARY KEY,
    notification_detail_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25),
    token TEXT NOT NULL,
    device_name VARCHAR,
    device_id VARCHAR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL
);

CREATE TABLE IF NOT EXISTS public.notification_logs (
    notification_id VARCHAR(25) PRIMARY KEY,
    notification_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25),
    title VARCHAR,
    message VARCHAR,
    image TEXT,
    type public.notification_type DEFAULT 'user',
    opened BOOLEAN DEFAULT FALSE,
    navigation_id VARCHAR(25),
    url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL
);

CREATE TABLE IF NOT EXISTS public.payment_methods (
    payment_method_id VARCHAR(25) PRIMARY KEY,
    payment_method_id_serial SERIAL UNIQUE,
    business_id VARCHAR(25) NOT NULL,
    name VARCHAR,
    phone_number VARCHAR,
    access_token VARCHAR,
    secret_key VARCHAR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (business_id) REFERENCES public.business(business_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.pre_orders (
    pre_order_id VARCHAR(25) PRIMARY KEY,
    pre_order_id_serial SERIAL UNIQUE,
    product_id VARCHAR(25) NOT NULL,
    product_name VARCHAR,
    product_images JSON,
    product_price NUMERIC(15,2) NOT NULL,
    business_id VARCHAR(25) NOT NULL,
    business_name VARCHAR,
    business_logo VARCHAR,
    user_id VARCHAR(25),
    delivery_id VARCHAR,
    delivery_name VARCHAR,
    delivery_address VARCHAR,
    delivery_phone_number VARCHAR,
    delivery_latitude NUMERIC(10,8),
    delivery_longitude NUMERIC(11,8),
    deleted_by_user BOOLEAN DEFAULT FALSE,
    deleted_by_business BOOLEAN DEFAULT FALSE,
    opened_by_user BOOLEAN DEFAULT FALSE,
    opened_by_business BOOLEAN DEFAULT FALSE,
    total_price NUMERIC(15,2) NOT NULL DEFAULT '0.00',
    quantity INTEGER NOT NULL DEFAULT 1,
    comment TEXT,
    payment_method VARCHAR,
    payment_id VARCHAR,
    payment_status public.payment_method DEFAULT 'pending',
    status public.preorder_status DEFAULT 'requested',
    ordered BOOLEAN DEFAULT FALSE,
    number_of_days INTEGER NOT NULL,
    pre_order_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (product_id) REFERENCES public.products(product_id) ON DELETE SET NULL,
    FOREIGN KEY (business_id) REFERENCES public.business(business_id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS public.privacies (
    privacy_id VARCHAR(25) PRIMARY KEY,
    privacy_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25) NOT NULL,
    blocked_user_id VARCHAR(25),
    blocked BOOLEAN,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.shares (
    share_id VARCHAR(25) PRIMARY KEY,
    share_id_serial SERIAL UNIQUE,
    ref_id VARCHAR(25) NOT NULL,
    user_id VARCHAR(25) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.sms_logs (
    sms_log_id VARCHAR(25) PRIMARY KEY,
    sms_log_id_serial SERIAL UNIQUE NOT NULL,
    sender_number VARCHAR(20) NOT NULL,
    recipient_number VARCHAR(20) NOT NULL,
    send_user_id VARCHAR(25) REFERENCES public.users(user_id) ON DELETE SET NULL ON UPDATE CASCADE,
    recipient_user_id VARCHAR(25) REFERENCES public.users(user_id) ON DELETE SET NULL ON UPDATE CASCADE,
    sms_message_sent JSONB,
    created_by VARCHAR(80) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(80) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.email_logs (
    email_log_id VARCHAR(25) PRIMARY KEY,
    email_log_id_serial SERIAL UNIQUE NOT NULL,
    sender_email VARCHAR(254) NOT NULL,
    recipient_email VARCHAR(254) NOT NULL,
    send_user_id VARCHAR(25) REFERENCES public.users(user_id) ON DELETE SET NULL ON UPDATE CASCADE,
    recipient_user_id VARCHAR(25) REFERENCES public.users(user_id) ON DELETE SET NULL ON UPDATE CASCADE,
    email_message_sent JSONB,
    created_by VARCHAR(80) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(80) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.transactions (
    transaction_id VARCHAR(25) PRIMARY KEY,
    transaction_id_serial SERIAL UNIQUE,
    sender_wallet_id VARCHAR(25) NOT NULL REFERENCES public.wallets(wallet_id) ON DELETE SET NULL,
    recipient_wallet_id VARCHAR(25) NOT NULL REFERENCES public.wallets(wallet_id) ON DELETE SET NULL,
    hash TEXT,
    amount NUMERIC(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'SLE',
    type public.transaction_type NOT NULL,
    status public.transaction_status NOT NULL DEFAULT 'pending',
    description TEXT,
    metadata JSONB,
    created_by VARCHAR(80) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(80) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.likes (
    like_id VARCHAR(25) PRIMARY KEY,
    like_id_serial SERIAL UNIQUE,
    ref_id VARCHAR(25),
    user_id VARCHAR(25) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.refund_requests (
    refund_request_id VARCHAR(25) PRIMARY KEY,
    refund_request_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25) NOT NULL,
    amount NUMERIC(15,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'SLE',
    status public.refund_request_status DEFAULT 'pending',
    external_transfer_details JSONB,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.refund_logs (
    refund_log_id VARCHAR(25) PRIMARY KEY,
    refund_log_id_serial SERIAL UNIQUE,
    refund_request_id VARCHAR(25) NOT NULL,
    user_id VARCHAR(25) NOT NULL,
    action VARCHAR(50) NOT NULL,
    details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (refund_request_id) REFERENCES public.refund_requests(refund_request_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.subscriptions (
    subscription_id VARCHAR(25) PRIMARY KEY,
    subscription_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25) NOT NULL,
    business_id VARCHAR(25) NOT NULL,
    plan VARCHAR(50) NOT NULL,
    status public.subscription_status NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (business_id) REFERENCES public.business(business_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.reports (
    report_id VARCHAR(25) PRIMARY KEY,
    report_id_serial SERIAL UNIQUE,
    user_id VARCHAR(25) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    content TEXT NOT NULL,
    pdf_url TEXT,
    image_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by VARCHAR(80) NOT NULL,
    updated_by VARCHAR(80) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Indexes for blogs
CREATE INDEX idx_blogs_user_id ON public.blogs(user_id);
CREATE INDEX idx_blogs_slug ON public.blogs(slug);
CREATE INDEX idx_blogs_created_at ON public.blogs(created_at);
CREATE INDEX idx_blogs_is_reel ON public.blogs(is_reel);
CREATE INDEX idx_blogs_views_count ON public.blogs(views_count);

-- Indexes for likes
CREATE INDEX idx_likes_user_id ON public.likes(user_id);
CREATE INDEX idx_likes_ref_id ON public.likes(ref_id);
CREATE INDEX idx_likes_created_at ON public.likes(created_at);

-- Indexes for public.roles
CREATE INDEX idx_roles_role_name ON public.roles(role_name);
CREATE INDEX idx_roles_created_at ON public.roles(created_at);

-- Indexes for public.users
CREATE INDEX idx_users_email ON public.users(email);
CREATE INDEX idx_users_status ON public.users(status);
CREATE INDEX idx_users_created_at ON public.users(created_at);

-- Indexes for public.users_stats
CREATE INDEX idx_users_stats_user_id ON public.users_stats(user_id);

-- Indexes for public.business_stats
CREATE INDEX idx_business_stats_business_id ON public.business_stats(business_id);

-- Indexes for public.user_roles
CREATE INDEX idx_user_roles_user_id ON public.user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON public.user_roles(role_id);
CREATE INDEX idx_user_roles_created_at ON public.user_roles(created_at);

-- Indexes for public.user_auths
CREATE INDEX idx_user_auths_user_id ON public.user_auths(user_id);
CREATE INDEX idx_user_auths_auth_provider ON public.user_auths(auth_provider);
CREATE INDEX idx_user_auths_created_at ON public.user_auths(created_at);

-- Indexes for public.user_settings
CREATE INDEX idx_user_settings_user_id ON public.user_settings(user_id);
CREATE INDEX idx_user_settings_created_at ON public.user_settings(created_at);

-- Indexes for public.encryption_keys
CREATE INDEX idx_encryption_keys_user_id ON public.encryption_keys(user_id);
CREATE INDEX idx_encryption_keys_created_at ON public.encryption_keys(created_at);

-- Indexes for public.business
CREATE INDEX idx_business_user_id ON public.business(user_id);
CREATE INDEX idx_business_email ON public.business(email);
CREATE INDEX idx_business_status ON public.business(status);
CREATE INDEX idx_business_created_at ON public.business(created_at);

-- Indexes for public.products
CREATE INDEX idx_products_business_id ON public.products(business_id);
CREATE INDEX idx_products_type ON public.products(type);
CREATE INDEX idx_products_category ON public.products(category);
CREATE INDEX idx_products_available ON public.products(available);
CREATE INDEX idx_products_created_at ON public.products(created_at);
CREATE INDEX idx_products_views_count ON public.products(views_count);

-- Indexes for public.orders
CREATE INDEX idx_orders_user_id ON public.orders(user_id);
CREATE INDEX idx_orders_business_id ON public.orders(business_id);
CREATE INDEX idx_orders_status ON public.orders(status);
CREATE INDEX idx_orders_payment_status ON public.orders(payment_status);
CREATE INDEX idx_orders_created_at ON public.orders(created_at);

-- Indexes for public.order_items
CREATE INDEX idx_order_items_order_id ON public.order_items(order_id);
CREATE INDEX idx_order_items_product_id ON public.order_items(product_id);
CREATE INDEX idx_order_items_created_at ON public.order_items(created_at);

-- Indexes for public.carts
CREATE INDEX idx_carts_user_id ON public.carts(user_id);
CREATE INDEX idx_carts_business_id ON public.carts(business_id);
CREATE INDEX idx_carts_created_at ON public.carts(created_at);

-- Indexes for public.cart_items
CREATE INDEX idx_cart_items_cart_id ON public.cart_items(cart_id);
CREATE INDEX idx_cart_items_product_id ON public.cart_items(product_id);
CREATE INDEX idx_cart_items_created_at ON public.cart_items(created_at);

-- Indexes for public.comments
CREATE INDEX idx_comments_user_id ON public.comments(user_id);
CREATE INDEX idx_comments_ref_id ON public.comments(ref_id);
CREATE INDEX idx_comments_created_at ON public.comments(created_at);

-- Indexes for public.views
CREATE INDEX idx_views_ref_id ON public.views(ref_id);
CREATE INDEX idx_views_user_id ON public.views(user_id);
CREATE INDEX idx_views_created_at ON public.views(created_at);

-- Indexes for public.pinned_blogs
CREATE INDEX idx_pinned_blogs_blog_id ON public.pinned_blogs(blog_id);
CREATE INDEX idx_pinned_blogs_user_id ON public.pinned_blogs(user_id);
CREATE INDEX idx_pinned_blogs_created_at ON public.pinned_blogs(created_at);

-- Indexes for public.pinned_products
CREATE INDEX idx_pinned_products_product_id ON public.pinned_products(product_id);
CREATE INDEX idx_pinned_products_user_id ON public.pinned_products(user_id);
CREATE INDEX idx_pinned_products_created_at ON public.pinned_products(created_at);

-- Indexes for public.contacts
CREATE INDEX idx_contacts_business_id ON public.contacts(business_id);
CREATE INDEX idx_contacts_created_at ON public.contacts(created_at);

-- Indexes for public.follows
CREATE INDEX idx_follows_follower_id ON public.follows(follower_id);
CREATE INDEX idx_follows_following_id ON public.follows(following_id);
CREATE INDEX idx_follows_created_at ON public.follows(created_at);

-- Indexes for public.notification_details
CREATE INDEX idx_notification_details_user_id ON public.notification_details(user_id);
CREATE INDEX idx_notification_details_device_id ON public.notification_details(device_id);
CREATE INDEX idx_notification_details_created_at ON public.notification_details(created_at);

-- Indexes for public.notification_logs
CREATE INDEX idx_notification_logs_user_id ON public.notification_logs(user_id);
CREATE INDEX idx_notification_logs_type ON public.notification_logs(type);
CREATE INDEX idx_notification_logs_opened ON public.notification_logs(opened);
CREATE INDEX idx_notification_logs_created_at ON public.notification_logs(created_at);

-- Indexes for public.payment_methods
CREATE INDEX idx_payment_methods_business_id ON public.payment_methods(business_id);
CREATE INDEX idx_payment_methods_created_at ON public.payment_methods(created_at);

-- Indexes for public.pre_orders
CREATE INDEX idx_pre_orders_product_id ON public.pre_orders(product_id);
CREATE INDEX idx_pre_orders_business_id ON public.pre_orders(business_id);
CREATE INDEX idx_pre_orders_user_id ON public.pre_orders(user_id);
CREATE INDEX idx_pre_orders_status ON public.pre_orders(status);
CREATE INDEX idx_pre_orders_payment_status ON public.pre_orders(payment_status);
CREATE INDEX idx_pre_orders_created_at ON public.pre_orders(created_at);

-- Indexes for public.privacies
CREATE INDEX idx_privacies_user_id ON public.privacies(user_id);
CREATE INDEX idx_privacies_blocked_user_id ON public.privacies(blocked_user_id);
CREATE INDEX idx_privacies_created_at ON public.privacies(created_at);

-- Indexes for public.shares
CREATE INDEX idx_shares_user_id ON public.shares(user_id);
CREATE INDEX idx_shares_ref_id ON public.shares(ref_id);
CREATE INDEX idx_shares_created_at ON public.shares(created_at);

-- Indexes for public.sms_logs
CREATE INDEX idx_sms_logs_send_user_id ON public.sms_logs(send_user_id);
CREATE INDEX idx_sms_logs_recipient_user_id ON public.sms_logs(recipient_user_id);
CREATE INDEX idx_sms_logs_created_at ON public.sms_logs(created_at);

-- Indexes for public.email_logs
CREATE INDEX idx_email_logs_send_user_id ON public.email_logs(send_user_id);
CREATE INDEX idx_email_logs_recipient_user_id ON public.email_logs(recipient_user_id);
CREATE INDEX idx_email_logs_created_at ON public.email_logs(created_at);

-- Indexes for public.wallets
CREATE INDEX idx_wallets_user_id ON public.wallets(user_id);
CREATE INDEX idx_wallets_is_active ON public.wallets(is_active);
CREATE INDEX idx_wallets_created_at ON public.wallets(created_at);

-- Indexes for public.transactions
CREATE INDEX idx_transactions_sender_wallet_id ON public.transactions(sender_wallet_id);
CREATE INDEX idx_transactions_recipient_wallet_id ON public.transactions(recipient_wallet_id);
CREATE INDEX idx_transactions_type ON public.transactions(type);
CREATE INDEX idx_transactions_status ON public.transactions(status);
CREATE INDEX idx_transactions_created_at ON public.transactions(created_at);

-- Indexes for refund_requests
CREATE INDEX idx_refund_requests_user_id ON public.refund_requests(user_id);
CREATE INDEX idx_refund_requests_status ON public.refund_requests(status);
CREATE INDEX idx_refund_requests_created_at ON public.refund_requests(created_at);

-- Indexes for refund_logs
CREATE INDEX idx_refund_logs_refund_request_id ON public.refund_logs(refund_request_id);
CREATE INDEX idx_refund_logs_user_id ON public.refund_logs(user_id);
CREATE INDEX idx_refund_logs_created_at ON public.refund_logs(created_at);

-- Indexes for subscriptions
CREATE INDEX idx_subscriptions_user_id ON public.subscriptions(user_id);
CREATE INDEX idx_subscriptions_business_id ON public.subscriptions(business_id);
CREATE INDEX idx_subscriptions_status ON public.subscriptions(status);
CREATE INDEX idx_subscriptions_created_at ON public.subscriptions(created_at);

-- Indexes for reports
CREATE INDEX idx_reports_user_id ON public.reports(user_id);
CREATE INDEX idx_reports_created_at ON public.reports(created_at);

-- Grant Access to role
GRANT USAGE ON SCHEMA public TO phinex;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO phinex;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO phinex;