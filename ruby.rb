require 'faraday'
require 'faraday_middleware'
require 'json'
require 'logger'

logger = Logger.new(STDOUT)
logger.level = Logger::INFO

unity = "unity.awstrp.net"
sonarqube_key = ENV['SONAR_KEY']
folio_headers = { 'Accept' => 'application/json', 'Content-Type' => 'application/json', 'Authorization' => "Basic #{sonarqube_key}" }

login = 'admin'

# Establish a connection using Faraday with middleware for handling redirects
conn = Faraday.new(url: "https://#{unity}") do |faraday|
  faraday.request :url_encoded        # Encode request parameters
  faraday.response :logger            # Log requests/responses (optional)
  faraday.use FaradayMiddleware::FollowRedirects # Automatically follow redirects
  faraday.adapter Faraday.default_adapter
end

# Perform a GET request to fetch business units
response = conn.get('/api/integration/hooks/businessUnits', nil, folio_headers)

# Check if the request was successful
if response.status == 200
  business_units = JSON.parse(response.body)
  business_units.each do |business_unit|
    key = business_unit.downcase.gsub(" ", "_")

    # Fix for specific business unit keys
    key = "dms2" if key == "dms"
    key = "gcas2" if key == "gcas"
    key = "xxx:" + key

    # Make a GET request to show the business unit
    response = conn.get("/sonarqube/api/views/show", { key: key }, folio_headers)

    if response.status == 200
      logger.info "Deleting: #{key}"

      # Add user permissions via POST request
      response = conn.post("/sonarqube/api/permissions/add_user", {
        login: login,
        permission: 'admin',
        projectKey: key
      }, folio_headers)

      # Delete the portfolio
      response = conn.post("/sonarqube/api/views/delete", { key: key }, folio_headers)

      if response.status == 204
        logger.info "Deleted #{business_unit} portfolio"
      else
        logger.error "#{response.status}: #{response.body} with #{url}"
      end
    end
  end
else
  logger.error "Error fetching business units: #{response.status} - #{response.body}"
end

# Get projects with valid app id
response = conn.get('/api/integration/hooks/projects/withValidAppId', nil, folio_headers)

if response.status == 200
  projects = JSON.parse(response.body)
  projects.each do |project|
    key = project['businessUnit'].downcase.gsub(" ", "_")

    # Fix for specific business unit keys
    key = "dms2" if key == "dms"
    key = "gcas2" if key == "gcas"
    key = "xxx:" + key

    # Add project to portfolio
    response = conn.post("/sonarqube/api/views/add_portfolio", {
      portfolio: key,
      reference: project['projectShortName']
    }, folio_headers)

    if response.status == 200
      logger.info "Added project #{project['projectShortName']} to #{key}"
    else
      logger.error "Error #{response.status}: #{response.body} when adding #{project['projectShortName']} to #{key}"
    end
  end
else
  logger.error "Error fetching projects: #{response.status} - #{response.body}"
end

# Refresh portfolios
business_units.each do |business_unit|
  key = business_unit.downcase.gsub(" ", "_")
  key = "dms2" if key == "dms"
  key = "gcas2" if key == "gcas"
  key = "xxx:" + key

  # Refresh the portfolio
  response = conn.post("/sonarqube/api/views/refresh", { key: key }, folio_headers)

  logger.info "Refreshed #{business_unit} portfolio"
end

['xxx:gcas2', 'xxx:gio'].each do |reference|
  # Add portfolios to GICO
  response = conn.post("/sonarqube/api/views/add_portfolio", {
    portfolio: 'gico',
    reference: reference
  }, folio_headers)

  if response.status != 200
    logger.warn "Error #{response.status}: #{response.body} when attempting to add #{reference} to GICO Technical Debt Management"
  end

  # Refresh the GICO portfolio
  response = conn.post("/sonarqube/api/views/refresh", { key: 'gico' }, folio_headers)

  logger.info "Refreshed GICO Technical Debt Management (key: gico) portfolio"
end
