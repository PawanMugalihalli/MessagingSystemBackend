-- USERS
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP
);

-- GROUPS
CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP,
    CONSTRAINT fk_groups_creator FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Index to filter/sort recent groups
CREATE INDEX idx_groups_created_at ON groups(created_at);
CREATE INDEX idx_groups_created_by ON groups(created_by);

-- GROUP MEMBERS
CREATE TABLE group_members (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    is_admin BOOLEAN DEFAULT false NOT NULL,
    joined_at TIMESTAMP,
    CONSTRAINT fk_group_member_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_group_member_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    UNIQUE (group_id, user_id)
);

-- Indexes for group membership filtering
CREATE INDEX idx_group_members_user_id ON group_members(user_id);
CREATE INDEX idx_group_members_group_id ON group_members(group_id);
CREATE INDEX idx_group_members_is_admin ON group_members(group_id, is_admin);
CREATE INDEX idx_group_members_joined_at ON group_members(joined_at);

-- GROUP MESSAGES
CREATE TABLE group_messages (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    CONSTRAINT fk_group_msg_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_group_msg_sender FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for fast retrieval and sorting
CREATE INDEX idx_group_messages_group_id ON group_messages(group_id);
CREATE INDEX idx_group_messages_sender_id ON group_messages(sender_id);
CREATE INDEX idx_group_messages_created_at ON group_messages(created_at);

-- DIRECT MESSAGES
CREATE TABLE direct_messages (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER NOT NULL,
    receiver_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    CONSTRAINT fk_dm_sender FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_dm_receiver FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for DM queries
CREATE INDEX idx_direct_messages_sender_id ON direct_messages(sender_id);
CREATE INDEX idx_direct_messages_receiver_id ON direct_messages(receiver_id);
CREATE INDEX idx_direct_messages_created_at ON direct_messages(created_at);
