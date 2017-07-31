DROP TABLE IF EXISTS Filters;
DROP TABLE IF EXISTS Dimensions;

CREATE TABLE Filters(
  filterId TEXT PRIMARY KEY,
  dataset TEXT,
  edition TEXT,
  version TEXT,
  state TEXT,
  filter JSONB NOT NULL
);

CREATE TABLE Dimensions(
  id SERIAL PRIMARY KEY,
  filterId TEXT,
  name TEXT,
  value TEXT
);
