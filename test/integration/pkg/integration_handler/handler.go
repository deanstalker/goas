package integration_handler

// @Title List all pets
// @ID listPets
// @Tag pets
// @Param limit query int false "How many items to return at one time (max 100)"
// @Success 200 object Pets "A paged array of pets"
// @Header 200 x-next string string "A link to the next page of responses"
// @Failure default object integration.Error "unexpected error"
// @Route /pets [get]
func listPets() {

}

// @Title Create a pet
// @ID createPets
// @Tag pets
// @Success 201 object "Null response"
// @Failure default object integration.Error "unexpected error"
// @Route /pets [post]
func createPets() {

}

// @Title Info for a specific pet
// @ID showPetById
// @Tag pets
// @Param petId path string true "The id of the pet to retrieve"
// @Success 200 object Pet "Expected response to a valid request"
// @Failure default object integration.Error "unexpected error"
func showPetById() {

}
