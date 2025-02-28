-- +goose Up
create table if not exists time_slots(
	id integer primary key,
	start_time datetime not null,
	end_time datetime not null,
	available boolean not null,
	title text,
	capacity integer not null,
	created_at datetime not null,
	updated_at datetime not null,
	deleted_at datetime
);

-- +goose Down
drop table if exists time_slots;