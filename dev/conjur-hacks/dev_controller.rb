# frozen_string_literal: true

class DevController < ApplicationController
  include BodyParser

  def index
    q = request.query_parameters

    case q[:action]
    when 'list_accounts'
      list_accounts(q)
    when 'retrieve_api_key'
      retrieve_api_key(q)
    else
      render(json: { error: 'must specify recognized action' })
    end
  end

  def list_accounts(_q)
    render(json: Account.list)
  end

  def retrieve_api_key(q)
    role = Role.first!(role_id: q[:role_id])
    render(plain: role.api_key)
  end
end
