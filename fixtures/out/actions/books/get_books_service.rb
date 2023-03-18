require "dry/monads"

module PetstoreApp
  module Actions
    module Books
      class GetBooksService
        include Dry::Monads[:result]

        def call(params)
          Success({})
        end
      end
    end
  end
end
