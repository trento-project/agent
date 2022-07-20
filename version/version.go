package version

// We exclude that variables from linting
// because we explicitly use that
// in the ldflags at build time
var Version string //nolint
var Flavor string  //nolint
