module TestApp
    class Routes < Hanami::Routes
        get "/books", to: "books.get_books"
        get "/books/:bookId", to: "books.get_book_by_id"
        
    end
end
