alter table gridfs_file
    modify id bigint auto_increment;

alter table gridfs_chunk
    modify id bigint auto_increment;

ALTER TABLE elastic
    ADD hosts JSON NULL AFTER password;
