# frozen_string_literal: true

# DevController is an add-on controller that powers the /dev endpoint
# on the Conjur instance used for development of the CLI.
# This endpoint allows us to easily manage common operations such as
# retrieving API keys by making HTTP requests to the
# Conjur endpoint.
class DevController < ApplicationController
  include BodyParser

  def params
    request.query_parameters
  end

  def index
    action = params[:action]
    raise ArgumentError, "'action' may not be empty" if action.blank?

    params[:account], params[:kind], params[:identifier] = params[:resource_id].split(":") unless params[:resource_id].blank?

    case action
    when 'create_account'
      create_account
    when 'create_secret'
      create_secret
    when 'get_secret'
      get_secret
    when 'load_policy'
      create_secret
    when 'list_accounts'
      list_accounts
    when 'destroy_account'
      destroy_account
    when 'retrieve_api_key'
      retrieve_api_key
    when 'purge'
      purge
    else
      raise ArgumentError, "'action' not recognized"
    end
  end

  def list_accounts
    controller = AccountsController.new
    controller.instance_variable_set("@current_user", Role["!:!:root"])
    controller.request = request
    controller.response = response
    controller.params = params

    render plain: controller.process(:index)
  end

  def destroy_account
    raise ArgumentError, "'id' may not be empty" if params[:id].blank?

    controller = AccountsController.new
    controller.instance_variable_set("@current_user", Role["!:!:root"])
    controller.request = request
    controller.response = response
    controller.params = params

    render plain: controller.process(:destroy)
  end

  def create_account
    raise ArgumentError, "'id' may not be empty" if params[:id].blank?

    controller = AccountsController.new
    controller.instance_variable_set("@current_user", Role["!:!:root"])
    controller.request = request
    controller.response = response
    controller.params = params

    render plain: controller.process(:create)
  end

  def retrieve_api_key
    role_id = params[:role_id]
    raise ArgumentError, "'role_id' may not be empty" if role_id.blank?

    role = Role.first!(role_id: role_id)
    render(plain: role.api_key)
  end

  def get_secret
    resource_id = params[:resource_id]

    raise ArgumentError, "'resource_id' may not be empty" if resource_id.blank?

    controller = SecretsController.new
    controller.instance_variable_set("@current_user", Role[params[:account] + ":user:admin"])
    controller.request = request
    controller.response = response
    controller.params = params

    render plain: controller.process(:show)
  end

  def create_secret
    resource_id = params[:resource_id]
    value = params[:value]

    raise ArgumentError, "'value' may not be empty" if value.blank?
    raise ArgumentError, "'resource_id' may not be empty" if resource_id.blank?

    request.headers['RAW_POST_DATA'] = value

    controller = SecretsController.new
    controller.instance_variable_set("@current_user", Role[params[:account] + ":user:admin"])
    controller.request = request
    controller.response = response
    controller.params = params

    render plain: controller.process(:create)
  end
end
