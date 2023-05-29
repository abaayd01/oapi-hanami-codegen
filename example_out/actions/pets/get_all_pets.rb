module API
  module Actions
    module Pets
      class GetAllPets < API::Action
        include Deps[service: "services.pets.get_all_pets"]
        params Contracts::GetAllPetsRequestContract

        def handle(request, response)
          service_result = service.call(request.params.to_h)

          if service_result.failure?
            raise StandardError
          end

          response_body_validation_result = Contracts::GetAllPetsResponseContract.new.call(service_result.value!.to_h)
          if response_body_validation_result.failure?
            raise BadResponseShapeError
          end

          response.body = response_body_validation_result.values.to_h.to_json
        end
      end
    end
  end
end
