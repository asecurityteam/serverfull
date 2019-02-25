// Package domain is a container of all of the domain types and interfaces
// that are used across multiple packages within the service.
//
// This package is also the container for all domain errors leveraged by the
// service. Each error here should represent a specific condition that needs to
// be communicated across interface boundaries.
//
// Generally speaking, this package contains no executable code. All elements are
// expected to be either pure data containers that have no associated methods or
// interface definitions that have no corresponding implementations in this package.
// The notable exception to this are the domain error types which are required to
// define a corresponding Error() method. Because these errors provide executable
// code they must also have corresponding tests. Only domain error types are allowed
// to deviate from the "no executable code" rule.
//
package domain
