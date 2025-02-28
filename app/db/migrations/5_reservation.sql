-- +goose Up
create table if not exists reservations(
	id integer primary key,
	timeslot_id integer not null,
	user_id integer not null,
	notes text,
	status text,
	created_at datetime not null,
	updated_at datetime not null,
	deleted_at datetime,
	foreign key(timeslot_id) references time_slots(id) on delete cascade
);

-- +goose Down
drop table if exists reservations;