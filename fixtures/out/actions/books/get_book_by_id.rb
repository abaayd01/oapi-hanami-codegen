module TestApp
  module Actions
    module Books
      class GetBookById < TestApp::Action
        def handle(request, response)
          request_validation_result = Contracts::GetBookByIdRequestContract.new.call(request.params.to_h)

          if request_validation_result.failure?
            Hanami.logger.error "request validation failure"
            Hanami.logger.error request_validation_result.errors.to_h
            response.body = "validation error"
            return
          end

          params = request_validation_result.values.to_h
          service_result = GetBookByIdService.new.call(params)

          if service_result.failure?
            Hanami.logger.error "service result failure"
            response.body = "internal server error"
            return
          end

          response_body_validation_result = Contracts::GetBookByIdResponseContract.new.call(service_result.value!.to_h)
          if response_body_validation_result.failure?
            Hanami.logger.error "response body validation error"
            Hanami.logger.error response_body_validation_result.errors.to_h
            response.body = "internal server error"
            return
          end

          response.body = response_body_validation_result.values.to_h.to_json
        end
      end
    end
  end
end
