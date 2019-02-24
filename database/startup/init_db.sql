DROP TABLE IF EXISTS server;
CREATE TABLE server
(
  id SERIAL NOT NULL PRIMARY KEY,
  guild_uid VARCHAR(20) NOT NULL UNIQUE,
  welcome_message VARCHAR(1900),
  rule_agreement VARCHAR(1900),
  veteran_rank INTEGER,
  veteran_role VARCHAR(20),
  bot_channel VARCHAR(20),
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  welcome_channel VARCHAR(20),
  starter_role VARCHAR(20),
  base_role VARCHAR(20)
);

DROP TABLE IF EXISTS channel;
CREATE TABLE channel
(
  id SERIAL NOT NULL PRIMARY KEY,
  server_id INTEGER NOT NULL REFERENCES server(Id) ON DELETE CASCADE,
  channel_uid VARCHAR(20) NOT NULL,
  bot_allowed BOOLEAN NOT NULL DEFAULT TRUE,
  move_pins BOOLEAN NOT NULL DEFAULT FALSE,
  move_text_pins BOOLEAN NOT NULL DEFAULT FALSE,
  delete_pin BOOLEAN NOT NULL DEFAULT FALSE,
  move_channel_uid TEXT CHECK (char_length(move_channel_uid) < 21)
);

DROP TABLE IF EXISTS role_group;
CREATE TABLE role_group
(
  id SERIAL NOT NULL PRIMARY KEY,
  server_id INTEGER NOT NULL REFERENCES server(id),
  name TEXT NOT NULL CHECK(char_length(name) <= 500),
  group_type INTEGER NOT NULL
);

DROP TABLE IF EXISTS role;
CREATE TABLE role
(
  id SERIAL NOT NULL PRIMARY KEY,
  server_id INTEGER REFERENCES server(Id) ON DELETE CASCADE,
  role_uid VARCHAR(20) NOT NULL UNIQUE,
  permission SMALLINT NOT NULL DEFAULT 2,
  confirmation_message VARCHAR CONSTRAINT role_confirmation_message_length CHECK (char_length(confirmation_message) <= 1900),
  confirmation_security_answer VARCHAR CONSTRAINT role_confirmation_security_answer_length CHECK (char_length(confirmation_security_answer) <= 1900),
  trigger TEXT CONSTRAINT role_trigger_length CHECK(char_length(Trigger) <= 100)
);

DROP TABLE IF EXISTS group_membership;
CREATE TABLE group_membership
(
  role_id  INTEGER NOT NULL REFERENCES role(id) ON DELETE CASCADE,
  role_group_id INTEGER NOT NULL REFERENCES role_group(id) ON DELETE CASCADE,
  CONSTRAINT group_membership_pkey PRIMARY KEY (role_id, role_group_id)
);

DROP TABLE IF EXISTS metric;
CREATE TABLE metric
(
  id SERIAL NOT NULL PRIMARY KEY,
  metric_type SMALLINT NOT NULL,
  data jsonb NOT NULL
);

DROP TABLE IF EXISTS poll;
CREATE TABLE poll
(
  id SERIAL NOT NULL PRIMARY KEY,
  title VARCHAR(100) NOT NULL,
  channel_id INTEGER NOT NULL REFERENCES channel(Id) ON DELETE CASCADE,
  user_uid VARCHAR(20) NOT NULL,
  message_uid VARCHAR(20),
  open BOOLEAN NOT NULL DEFAULT TRUE
);

DROP TABLE IF EXISTS poll_option;
CREATE TABLE poll_option
(
  id SERIAL NOT NULL PRIMARY KEY,
  poll_id INTEGER REFERENCES poll(Id) ON DELETE CASCADE,
  reaction_uid VARCHAR(20) NOT NULL,
  reaction_name VARCHAR(30) NOT NULL,
  description VARCHAR(200) NOT NULL,
  votes INTEGER NOT NULL DEFAULT 0
);

DROP TABLE IF EXISTS scheduled_operation;
CREATE TABLE scheduled_operation
(
  id SERIAL NOT NULL PRIMARY KEY,
  server_id INTEGER NOT NULL REFERENCES server(id) ON DELETE CASCADE,
  scheduler_type INTEGER NOT NULL,
  planned_execution_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  execution_interval INTERVAL NOT NULL
);

DROP TABLE IF EXISTS channel_rotation;
CREATE TABLE channel_rotation
(
  operation_id INTEGER NOT NULL PRIMARY KEY REFERENCES scheduled_operation(id) ON DELETE CASCADE,
  current_channel_uid VARCHAR(20) NOT NULL,
  channel_uids VARCHAR(1000) NOT NULL
);

DROP TABLE IF EXISTS user_profile;
CREATE TABLE user_profile
(
  id SERIAL NOT NULL PRIMARY KEY,
  user_uid VARCHAR(20) NOT NULL UNIQUE
);

DROP TABLE IF EXISTS user_server_rank;
CREATE TABLE user_server_rank
(
  id SERIAL NOT NULL PRIMARY KEY,
  server_id INTEGER NOT NULL REFERENCES server(id) ON DELETE CASCADE,
  user_id INTEGER NOT NULL REFERENCES user_profile(id) ON DELETE CASCADE,
  rank INTEGER NOT NULL DEFAULT 0,
  message_sent BOOLEAN NOT NULL DEFAULT false
);
