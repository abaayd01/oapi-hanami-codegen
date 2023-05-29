require "dry/monads"

module API
  module Services
    module Pets
      class CreatePet
        include Dry::Monads[:result]

        def call(params)
          Success({})
        end
      end
    end
  end
end
