## API development considerations

There is no predefined RFC/standard of how RESTful APIs should be designed, however we follow the best practices and try to adhere to the following patterns.

URLs should not have trailing slash at the end if there is no hiearchy of operations.

GOOD `/quotes`
BAD `/quotes/`

GOOD `/quotes/:id`
BAD `/quotes/:id/`

Given a `resource` named `Quote` the following URLs and their respective http `METHOD` would be used.

* `GET /quotes`
    * Get all quotes available, notice the plural form, the response can be paginated or not.

* `GET /quotes/:id`
    * Get a specific quote by `:id` where `:id` is most likely an `uuid`.

* `GET /quotes?date_from=xyz`
    * Search/Filter quotes. This is the same as performing a simple `GET /quotes` but the results should be the ones we are looking for, the response can be paginated or not.

* `POST /quotes`
    * Allows the creation of a `Quote`, the body of the request would contain the attributes of the `Quote` object, no need to pass an `:id`.

* `PATCH /quotes/:id`
    * Allows the modification of a `Quote` by the `:id`, the body of the request would contain the attributes of the `Quote` object to be update with.

* `DELETE /quotes/:id`
    * Allows the deletion of a `Quote` by the `:id`

There are additional cases where we might want to add additional details in the URLs, for instance if we want to download a PDF file for a given `Quote` the URL should look like this `GET /quotes/:id/pdf`


## API status Codes

The backend exposes functionality via a RESTful API. The following error codes describe the expected behaviour of the backend.

The response body is in `JSON` format.

* `200` OK
    * The request you have made was successful, the response might contain relevant data to the request, such as the `Object` to retrieve or the `Object` which was created.
* `400` BAD REQUEST
    * Making an API call with the wrong parameters/missing parameters will result in a `400`, the response body will contain the relevant data related to the request.
* `401` UNAUTHORIZED
    * The `frontend` should redirect the `User` for a `Login`. This can happen if the user's session has expired or simply trying to make an API request while not being authenticated.
* `403` PERMISSION DENIED
    * The API call requires you to have permissions on the given `Resource`. Should not really happen in real world.
* `405` METHOD NOT ALLOWED
    * The API endpoint does not allow to be called with the specific HTTP `METHOD`, ensure you are using the correct `METHOD` such as `GET`, `POST` etc.
* `500` INTERNAL SERVER ERROR
    * The server was not able to deal with the request for various reasons... The `frontend` should show an `Error` popup or similar.
