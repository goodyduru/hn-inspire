DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS flags;
DROP TABLE IF EXISTS favorites;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id          serial primary key,
    email       varchar(254) unique,
    username    varchar(150) not null unique,
    password    varchar(128) not null,
    is_admin    boolean default false,
    date_joined timestamp with time zone default now(),
    karma       integer default 0,
    about       text
);

CREATE TABLE posts (
    id          serial primary key,
    title       varchar(255) not null,
    url         varchar(255) not null,
    text        text not null,
    votes       integer default 0,
    is_flagged  boolean default false,
    author_id   integer references users on delete set null,
    parent_id   integer references posts on delete set null,
    created_at  timestamp with time zone default now(),
    updated_at  timestamp with time zone default now()
);

CREATE TABLE votes (
    user_id     integer references users on delete cascade,
    post_id     integer references posts on delete cascade,
    amount      integer,
    PRIMARY KEY (user_id, post_id)
);

CREATE TABLE flags (
    user_id     integer references users on delete cascade,
    post_id     integer references posts on delete cascade,
    PRIMARY KEY (user_id, post_id)
);

CREATE TABLE favorites (
    user_id     integer references users on delete cascade,
    post_id     integer references posts on delete cascade,
    PRIMARY KEY (user_id, post_id)
)
