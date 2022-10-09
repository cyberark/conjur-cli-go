class DevController < ApplicationController
  include BodyParser

  def index
    query_params = request.query_parameters

    case query_params[:action]
    when 'list_accounts'
      list_accounts
    when 'retrieve_api_key'
      retrieve_api_key(query_params)
    else
      render(json: { error: 'must specify recognized action' })
    end
  end

  def list_accounts
    render(json: Account.list)
  end

  def retrieve_api_key(query_params)
    role = Role.first!(role_id: query_params[:role_id])
    render(plain: role.api_key)
  end
end
