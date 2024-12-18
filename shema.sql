DROP TABLE IF EXISTS news,shortnews;
CREATE TABLE news (
  id BIGSERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  content TEXT ,
  preview TEXT ,
  published BIGINT,
  link TEXT NOT NULL UNIQUE 
);

