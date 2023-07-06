alter table gridfs_file
    modify id bigint auto_increment;

alter table gridfs_chunk
    modify id bigint auto_increment;

alter table minion
    add clam TINYINT(1) default 0 not null;
