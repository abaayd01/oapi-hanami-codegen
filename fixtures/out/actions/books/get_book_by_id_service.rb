require "dry/monads"

module TestApp
  module Actions
    module Books
      class GetBookByIdService
        include Dry::Monads[:result]

        def call(params)
          Success({})
        end
      end
    end
  end
end
