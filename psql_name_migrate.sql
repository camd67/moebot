-- Channel
ALTER TABLE channel
RENAME serverid TO server_id;

ALTER TABLE channel
RENAME channeluid TO channel_uid;

ALTER TABLE channel
RENAME botallowed TO bot_allowed;

ALTER TABLE channel
RENAME movepins TO move_pins;

ALTER TABLE channel
RENAME movetextpins TO move_text_pins;



-- Metric
ALTER TABLE Metric
RENAME type TO metric_type;



-- Poll
ALTER TABLE poll
RENAME channelid TO channel_id;

ALTER TABLE poll
RENAME useruid TO user_uid;

ALTER TABLE poll
RENAME messageuid TO message_uid;



-- Poll Option
ALTER TABLE poll_option
RENAME pollid TO poll_id;

ALTER TABLE poll_option
RENAME reactionid TO reaction_id;

ALTER TABLE poll_option
RENAME reactionname TO reaction_name;



-- Raffle Entry
ALTER TABLE raffle_entry
RENAME guilduid TO guild_uid;

ALTER TABLE raffle_entry
RENAME useruid TO user_uid;

ALTER TABLE raffle_entry
RENAME raffletype TO raffle_type;

ALTER TABLE raffle_entry
RENAME ticketcount TO ticket_count;

ALTER TABLE raffle_entry
RENAME raffledata TO raffle_data;

ALTER TABLE raffle_entry
RENAME lastticketupdate TO last_ticket_update;



-- Role
ALTER TABLE role
RENAME serverid TO server_id;

ALTER TABLE role
RENAME roleuid TO role_uid;

ALTER TABLE role
RENAME confirmationmessage TO confirmation_message;

ALTER TABLE role
RENAME confirmationsecurityanswer TO confirmation_security_answer;

ALTER TABLE role
DROP COLUMN groupid;



-- Role Group
ALTER TABLE role_group
RENAME serverid TO server_id;

ALTER TABLE role_group
RENAME type TO group_type;



-- Group Membership
ALTER TABLE group_membership
RENAME group_id TO role_group_id;


-- Server
ALTER TABLE server
RENAME guilduid TO guild_uid;

ALTER TABLE server
RENAME welcomemessage TO welcome_message;

ALTER TABLE server
RENAME ruleagreement TO rule_agreement;

ALTER TABLE server
RENAME veteranrank TO veteran_rank;

ALTER TABLE server
RENAME veteranrole TO veteran_role;

ALTER TABLE server
RENAME botchannel TO bot_channel;

ALTER TABLE server
RENAME welcomechannel TO welcome_channel;

ALTER TABLE server
RENAME starterrole TO starter_role;

ALTER TABLE server
RENAME baserole TO base_role;



-- User Profile
ALTER TABLE user_profile
RENAME useruid TO user_uid;



-- User Server Rank
ALTER TABLE user_server_rank
RENAME serverid TO server_id;

ALTER TABLE user_server_rank
RENAME userid TO user_id;

ALTER TABLE user_server_rank
RENAME messagesent TO message_sent;



-- Scheduled Operation
ALTER TABLE scheduled_operation
RENAME type TO scheduler_type;