require "dry/monads"

module API
  module Services
    module Books
      class GetBooks
        include Dry::Monads[:result]

        def call(params)
          Success({})
        end
      end
    end
  end
end
