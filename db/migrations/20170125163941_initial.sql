
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

DROP TABLE IF EXISTS tree;
CREATE TABLE tree (
    id serial PRIMARY KEY,
    left_key integer,
    right_key integer,
    value character varying(30)
);


INSERT INTO tree (left_key, right_key, value) VALUES
(1,18,'animals'),
(2,7,'insects'),
(3,4,'bees'),
(5,6,'flies'),
(8,13,'mammals'),
(9,10,'dogs'),
(11,12,'cats'),
(14,17,'fish'),
(15,16,'sharks')
;


CREATE INDEX left_key_idx ON tree USING btree (left_key);
CREATE INDEX right_key_idx ON tree USING btree (right_key);
CREATE INDEX value_idx ON tree USING btree (value);
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE tree;
