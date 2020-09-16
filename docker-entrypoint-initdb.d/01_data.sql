INSERT INTO users(login, password, roles)
VALUES
       ('admin', '$2a$10$ctPFhgJh.YIE21AA0OGl5er3p9f3XsAwkmTXnP2I7BxCpQbr1QAg2', '{"USER"}'),
       ('user', '$2a$10$ctPFhgJh.YIE21AA0OGl5er3p9f3XsAwkmTXnP2I7BxCpQbr1QAg2', '{"USER"}'),
       ('service', '$2a$10$ctPFhgJh.YIE21AA0OGl5er3p9f3XsAwkmTXnP2I7BxCpQbr1QAg2', '{"SERVICE"}');

INSERT INTO cards (number, balance, issuer, holder, user_id, status)
VALUES ('1234', 1000000, 'Visa', 'user1', 1, 'ACTIVE'),
       ('2345', 1000000, 'Visa', 'user2', 2, 'ACTIVE'),
       ('2345', 1000000, 'MasterCard', 'user2', 2, 'ACTIVE');

INSERT INTO tokens(id, user_id)
VALUES
       ('88916179-1692-418b-871b-38d97c90b12a', '1'),
       ('88916179-1692-418b-871b-38d97c90b12h', '2');


