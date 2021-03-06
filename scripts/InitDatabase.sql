DROP TABLE IF EXISTS Filters;
DROP TABLE IF EXISTS Dimensions;
DROP TABLE IF EXISTS Downloads;

CREATE TABLE Filters(
  filterJobId TEXT PRIMARY KEY,
  instanceId TEXT,
  versionId TEXT,
  versionHref TEXT,
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

ALTER TABLE Dimensions
  ADD CONSTRAINT filterJobDimensionOption
  UNIQUE (filterJobId, name, option);


ALTER TABLE Downloads
  ADD CONSTRAINT filterJobDownloadURL
  UNIQUE (filterJobId, type);
