DROP SEQUENCE IF EXISTS "public"."users_id_seq";
CREATE SEQUENCE "public"."users_id_seq"
    INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1;

DROP TABLE IF EXISTS "public"."users";
CREATE TABLE "public"."users"
(
    "id"       int4                                       NOT NULL DEFAULT nextval('users_id_seq'::regclass),
    "name"     varchar(50) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'YOK':: character varying,
    "surname"  varchar(50) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'YOK':: character varying,
    "email"    varchar(62) COLLATE "pg_catalog"."default" NOT NULL DEFAULT '':: character varying,
    "password" varchar(64) COLLATE "pg_catalog"."default" NOT NULL DEFAULT '':: character varying
)
;

INSERT INTO "public"."users" VALUES (1, 'ELON', 'MUSK', 'elon.musk@tesla.com', 'QJe349yb1rRh3S+Ga8PGrFQuhvOWmdTSGs2nrWdTdFM=');
/*
 LOGIN
 password: test
*/


ALTER SEQUENCE "public"."users_id_seq"
    OWNED BY "public"."users"."id";
SELECT setval('"public"."users_id_seq"', 2, true);


ALTER TABLE "public"."users"
    ADD CONSTRAINT "users_pkey" PRIMARY KEY ("id");