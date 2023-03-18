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
  optional(:avatar).value(:hash) do
  optional(:profile_image_url).value(:string)
  optional(:id).value(:integer)
  end
  required(:reviews).array(:hash) do
  optional(:rating).value(:integer)
  required(:text).value(:string)
  required(:user).value(:hash) do
  optional(:name).value(:string)
  optional(:id).value(:string)
  end
  end
  required(:title).value(:string)
  required(:author).value(:string)
        end
      end
      
    end
  end
end
