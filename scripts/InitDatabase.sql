DROP TABLE IF EXISTS Filters;
DROP TABLE IF EXISTS Dimensions;
DROP TABLE IF EXISTS Downloads;

CREATE TABLE Filters(
  filterJobId TEXT PRIMARY KEY,
  datasetFilterId TEXT,
  state TEXT
);

CREATE TABLE Dimensions(
  id SERIAL PRIMARY KEY,
  filterJobId TEXT,
  name TEXT,
  option TEXT
);

CREATE TABLE Downloads(
  id SERIAL PRIMARY KEY,
  filterJobId TEXT,
  size TEXT,
  type TEXT,
  url TEXT
);
