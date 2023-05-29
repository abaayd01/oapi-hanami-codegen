require "dry/validation"

module API
  module Actions
    module Schemas
      
      Owner = Dry::Schema.Params do
  optional(:age).value(:integer)
  optional(:id).value(:string)
  optional(:name).value(:string)
      end
      
      Pet = Dry::Schema.Params do
  optional(:id).value(:integer)
  required(:name).value(:string)
  optional(:nicknames).array(:string)
  optional(:owners).array(Schemas::Owner)
      end
      
    end
  end
end
