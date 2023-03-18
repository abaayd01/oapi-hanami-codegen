require "dry/monads"

module PetstoreApp
  module Actions
    module Pets
      class GetAllPetsService
        include Dry::Monads[:result]

        def call(params)
          Success({})
        end
      end
    end
  end
end
