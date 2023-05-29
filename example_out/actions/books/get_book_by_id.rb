module API
  module Actions
    module Books
      class GetBookById < API::Action
        include Deps[service: "services.books.get_book_by_id"]
        params Contracts::GetBookByIdRequestContract

        def handle(request, response)
          service_result = service.call(request.params.to_h)

          if service_result.failure?
            raise StandardError
          end

          response_body_validation_result = Contracts::GetBookByIdResponseContract.new.call(service_result.value!.to_h)
          if response_body_validation_result.failure?
            raise BadResponseShapeError
          end

          response.body = response_body_validation_result.values.to_h.to_json
        end
      end
    end
  end
end
