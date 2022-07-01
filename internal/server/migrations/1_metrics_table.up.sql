create table metrics (
    id character varying not null,
    type character varying not null,
    delta bigint default 0,
    gauge double precision default 0,
    unique (id, type)
);

create index metrics_id_idx ON metrics (id);
create index metrics_mtype_idx ON metrics (type);
