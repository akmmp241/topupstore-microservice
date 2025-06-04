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

alter table users
    add column name varchar(255) not null;
alter table users
    add column phone_number varchar(255) not null;

create unique index users_email_uindex
    on users (email);

create index users_phone_number_index
    on users (phone_number);

create table categories
(
    id         bigint auto_increment primary key,
    ref_id     varchar(255)                        not null,
    name       varchar(255)                        not null,
    created_at timestamp default CURRENT_TIMESTAMP not null,
    updated_at timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    engine = innodb;

create table operators
(
    id          bigint auto_increment primary key,
    ref_id      varchar(255)                           not null,
    category_id bigint                                 not null,
    name        varchar(255)                           not null,
    slug        varchar(255)                           not null,
    image_url   varchar(255) default null              null,
    description varchar(255) default null              null,
    created_at  timestamp    default CURRENT_TIMESTAMP not null,
    updated_at  timestamp    default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,

    constraint operators_category_id_foreign
        foreign key (category_id) references categories (id) on delete cascade on update cascade
)
    engine = innodb;

create table product_types
(
    id          bigint auto_increment primary key,
    ref_id      varchar(255)                        not null,
    operator_id bigint                              not null,
    name        varchar(255)                        not null,
    format_form text                                not null,
    created_at  timestamp default CURRENT_TIMESTAMP not null,
    updated_at  timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,

    constraint product_types_operator_id_foreign
        foreign key (operator_id) references operators (id) on delete cascade on update cascade
)
    engine = innodb;

create table products
(
    id              bigint auto_increment primary key,
    ref_id          varchar(255)                           not null,
    product_type_id bigint                                 not null,
    name            varchar(255)                           not null,
    description     text                                   not null,
    image_url       varchar(255) default null              null,
    price           int                                    not null,
    created_at      timestamp    default CURRENT_TIMESTAMP not null,
    updated_at      timestamp    default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,

    constraint products_product_type_id_foreign
        foreign key (product_type_id) references product_types (id) on delete cascade on update cascade
)
    engine = innodb;

-- Insert categories
INSERT INTO categories (ref_id, name)
VALUES ('CAT001', 'Mobile Prepaid'),
       ('CAT002', 'Internet Services'),
       ('CAT003', 'Digital TV');

-- Insert operators
INSERT INTO operators (ref_id, category_id, name, slug, image_url, description)
VALUES ('OP001', 1, 'TelcoOne', 'telco-one', 'https://example.com/telco1.jpg', 'Leading mobile operator'),
       ('OP002', 1, 'MobileMax', 'mobile-max', 'https://example.com/mobile2.jpg', 'Best coverage network'),
       ('OP003', 2, 'WebNet', 'web-net', 'https://example.com/webnet.jpg', 'Fast internet provider');

-- Insert product types
INSERT INTO product_types (ref_id, operator_id, name, format_form)
VALUES ('PT001', 1, 'Prepaid Credit', '{"phone_number": "string", "amount": "number"}'),
       ('PT002', 2, 'Data Package', '{"phone_number": "string", "package_id": "string"}'),
       ('PT003', 3, 'Internet Package', '{"customer_id": "string", "plan_type": "string"}');

-- Insert products
INSERT INTO products (ref_id, product_type_id, name, description, image_url, price)
VALUES ('PRD001', 1, '10 USD Credit', 'Phone credit worth 10 USD', 'https://example.com/credit10.jpg', 1000),
       ('PRD002', 1, '20 USD Credit', 'Phone credit worth 20 USD', 'https://example.com/credit20.jpg', 2000),
       ('PRD003', 2, '5GB Data Pack', '5GB data valid for 30 days', 'https://example.com/data5gb.jpg', 1500),
       ('PRD004', 3, 'Fiber 100Mbps', '100Mbps fiber internet package', 'https://example.com/fiber100.jpg', 5000);

create table orders
(
    id                   varchar(255) primary key,
    buyer_id             bigint                                 not null,
    buyer_email          varchar(255) default null              null,
    buyer_phone          varchar(255)                           not null,
    product_id           bigint                                 not null,
    product_name         varchar(255)                           not null,
    destination          varchar(255)                           not null,
    server_id            varchar(255)                           not null,
    payment_method_id    varchar(255)                           not null,
    payment_method_name  varchar(255)                           not null,
    total_product_amount int                                    not null,
    service_charge       int                                    not null,
    total_amount         int                                    not null,
    status               varchar(50)                            not null,
    created_at           timestamp    default CURRENT_TIMESTAMP not null,
    updated_at           timestamp    default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    engine = innodb;

alter table orders modify column buyer_id bigint default null null;

alter table orders modify column buyer_email varchar(255) not null;
alter table orders modify column buyer_phone varchar(255) default null null;
alter table orders add column failure_code varchar(50) default null null after status;