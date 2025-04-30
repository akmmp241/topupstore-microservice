create table users
(
    id                       bigint auto_increment
        primary key,
    email                    varchar(255)                           not null,
    password                 varchar(255)                           not null,
    email_verification_token varchar(255) default null              null,
    email_verified_at        timestamp    default null              null,
    created_at               timestamp    default CURRENT_TIMESTAMP not null,
    updated_at               timestamp    default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    engine = innodb;

alter table users add column name varchar(255) not null ;
alter table users add column phone_number varchar(255) not null ;

create unique index users_email_uindex
    on users (email);

create index users_phone_number_index
    on users (phone_number);

SELECT id, name, email, phone_number, email_verification_token, email_verified_at, created_at, updated_at FROM users WHERE email = 'example@gmail.com';