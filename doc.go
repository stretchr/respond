// Package respond provides a simple and consistent way of writing
// data API responses when building web services.
//
//     // write some data
//     respond.With{Data:obj}.To(w,r)
//
//     // with specific status
//     respond.With{
//       Data:obj,
//       Status:http.StatusCreated,
//     }.To(w,r)
//
package respond
