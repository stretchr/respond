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
//     // adding a default header
//     respond.DefaultOptions.Headers.Set("X-App-Version", "1.0")
//
//     // adding a specific header
//     respond.With{
//       Data:obj,
//       Status:http.StatusCreated,
//       Headers:map[string][]string{"X-RateLimit-Remaining": []string{remaining}},
//     }.To(w,r)
//
package respond
