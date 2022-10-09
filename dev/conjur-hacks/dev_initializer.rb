Rails.application.configure do
  # Extend routes by monkey-patching the Rails.application.routes.draw method
  old_draw = Rails.application.routes.method(:draw)
  new_draw = lambda { |*args, &block|
    old_draw.call do
      scope format: false do
        # Here's where we extend the router :)

        get '/dev' => 'dev#index'
      end
    end

    old_draw.call(*args, &block)
  }
  Rails.application.routes.define_singleton_method :draw, new_draw

  # Reload dev controller if it changes between requests
  dev_controller_path = Rails.root.join('app/controllers/dev_controller.rb');
  dev_reloader = ActiveSupport::FileUpdateChecker.new([dev_controller_path]) do
    load Rails.root.join(dev_controller_path)
  end
  ActionDispatch::Callbacks.before do
    dev_reloader.execute_if_updated
  end

  # Allow /dev routes to be accessed without authn/authz
  config.after_initialize do
    Rails
    .application
    .middleware
    .find { |m|
      m.name == 'Conjur::Rack::Authenticator' 
    }
    .args[0][:except]
    .append(%r{^/dev})
  end
end
