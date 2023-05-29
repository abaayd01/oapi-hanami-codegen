module PetstoreApp
  class Routes < Hanami::Routes
    slice :api, at: "/api" do
      get "/books", to: "books.get_books"
      get "/books/:bookId", to: "books.get_book_by_id"
      get "/pets", to: "pets.get_all_pets"
      post "/pets", to: "pets.create_pet"
      get "/pets/:petId", to: "pets.get_pet_by_id"
      
    end
  end
end
