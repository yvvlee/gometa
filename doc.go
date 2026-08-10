// Package gometa provides strongly typed declaration metadata written as
// ordinary Go code.
//
// Metadata is exposed through MetadataProvider. No reflection or code
// generation is needed at runtime. The companion analyzer checks that field,
// method, parameter, and result references still match their Go declarations.
package gometa
