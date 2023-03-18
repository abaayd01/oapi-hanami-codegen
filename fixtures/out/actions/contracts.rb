require "dry/validation"

module TestApp
  module Actions
    module Contracts
      class GetBooksRequestContract < Dry::Validation::Contract
        params do
        end
      end
      
      class GetBooksResponseContract < Dry::Validation::Contract
        params do
  optional(:books).array(:hash) do
  optional(:author).value(:string)
  optional(:title).value(:string)
  end
        end
      end
      
      class GetBookByIdRequestContract < Dry::Validation::Contract
        params do
        end
      end
      
      class GetBookByIdResponseContract < Dry::Validation::Contract
        params do
  required(:author).value(:string)
  optional(:avatar).value(:hash) do
  optional(:id).value(:integer)
  optional(:profile_image_url).value(:string)
  end
  required(:reviews).array(:hash) do
  optional(:rating).value(:integer)
  required(:text).value(:string)
  required(:user).value(:hash) do
  optional(:id).value(:string)
  optional(:name).value(:string)
  end
  end
  required(:title).value(:string)
        end
      end
      
    end
  end
end
