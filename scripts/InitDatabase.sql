DROP TABLE IF EXISTS Filters;
DROP TABLE IF EXISTS Dimensions;

CREATE TABLE Filters(
  filterJobId TEXT PRIMARY KEY,
  datasetFilterId TEXT,
  state TEXT
);

CREATE TABLE Dimensions(
  id SERIAL PRIMARY KEY,
  filterJobId TEXT,
  name TEXT,
  value TEXT
);
