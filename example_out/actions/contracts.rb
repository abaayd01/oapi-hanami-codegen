require "dry/validation"

module API
  module Actions
    module Contracts
      class GetBooksRequestContract < Hanami::Action::Params
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
      
      class GetBookByIdRequestContract < Hanami::Action::Params
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
      
      class GetAllPetsRequestContract < Hanami::Action::Params
        params do
  required(:page).value(:integer)
  optional(:q).value(:string)
        end
      end
      
      class GetAllPetsResponseContract < Dry::Validation::Contract
        params do
  optional(:age).value(:integer)
  optional(:name).value(:string)
        end
      end
      
      class CreatePetRequestContract < Hanami::Action::Params
        params do
  optional(:age).value(:integer)
  optional(:name).value(:string)
        end
      end
      
      class CreatePetResponseContract < Dry::Validation::Contract
        params do
  optional(:age).value(:integer)
  optional(:name).value(:string)
        end
      end
      
      class GetPetByIdRequestContract < Hanami::Action::Params
        params do
  required(:pet_id).value(:integer)
        end
      end
      
      class GetPetByIdResponseContract < Dry::Validation::Contract
        params do
  optional(:id).value(:integer)
  required(:name).value(:string)
  optional(:nicknames).array(:string)
  optional(:owners).array(Schemas::Owner)
        end
      end
      
    end
  end
end
