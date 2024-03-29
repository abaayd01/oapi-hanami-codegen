# auto_register: false
# frozen_string_literal: true

require "hanami/action"

module PetstoreApp
  class BaseAction < Hanami::Action
    format :json

    before :validate_request_params

    ForbiddenError = Class.new(StandardError)
    NotFoundError = Class.new(StandardError)
    BadResponseShapeError = Class.new(StandardError)

    # Exception handling
    config.handle_exception ForbiddenError => :handle_forbidden
    config.handle_exception StandardError => :handle_standard_error
    config.handle_exception NotFoundError => :handle_not_found
    config.handle_exception BadResponseShapeError => :handle_standard_error

    private

    def validate_request_params(req, res)
      halt 422, { errors: req.params.errors }.to_json unless req.params.valid?
    end

    def handle_forbidden(_req, res, exception)
      res.status = 403
      res.body = { error: "Forbidden" }.to_json
    end

    def handle_not_found(_req, res, _exception)
      res.status = 404
      res.body = { error: "Not found" }.to_json
    end

    def handle_standard_error(_req, res, exception)
      # todo: include sentry in deps
      # sentry.capture_exception(exception)
      res.status = 500
      res.body = { error: "Something went wrong processing your request" }.to_json
    end
  end
end
