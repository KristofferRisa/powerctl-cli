package api

// GraphQL queries for Tibber API

// QueryHomes fetches all homes with their details
const QueryHomes = `{
  viewer {
    homes {
      id
      appNickname
      size
      type
      numberOfResidents
      primaryHeatingSource
      hasVentilationSystem
      mainFuseSize
      features {
        realTimeConsumptionEnabled
      }
      address {
        address1
        address2
        address3
        postalCode
        city
        country
        latitude
        longitude
      }
    }
  }
}`

// QueryPrices fetches current and upcoming prices
const QueryPrices = `{
  viewer {
    homes {
      id
      currentSubscription {
        priceInfo {
          current {
            total
            energy
            tax
            startsAt
            level
            currency
          }
          today {
            total
            energy
            tax
            startsAt
            level
            currency
          }
          tomorrow {
            total
            energy
            tax
            startsAt
            level
            currency
          }
        }
      }
    }
  }
}`

// SubscriptionLiveMeasurement is the GraphQL subscription for real-time data
const SubscriptionLiveMeasurement = `subscription($homeId: ID!) {
  liveMeasurement(homeId: $homeId) {
    timestamp
    power
    powerProduction
    accumulatedConsumption
    accumulatedProduction
    accumulatedCost
    accumulatedReward
    minPower
    maxPower
    averagePower
    voltagePhase1
    voltagePhase2
    voltagePhase3
    currentL1
    currentL2
	currentL3
    currency
  }
}`

// QueryConsumptionHome fetches consumption history for a specific home directly.
// $before is a cursor from a previous page's pageInfo.startCursor; omit it to
// get the newest Nodes.
const QueryConsumptionHome = `query($homeId: ID!, $resolution: EnergyResolution!, $last: Int!, $before: String) {
  viewer {
    home(id: $homeId) {
      id
      consumption(resolution: $resolution, last: $last, before: $before) {
        pageInfo {
          startCursor
          hasPreviousPage
        }
        nodes {
          from
          to
          consumption
          cost
          unitPrice
          unitPriceVAT
          currency
        }
      }
    }
  }
}`

// QueryConsumptionHomes fetches consumption history for all homes (used as fallback)
const QueryConsumptionHomes = `query($resolution: EnergyResolution!, $last: Int!, $before: String) {
  viewer {
    homes {
      id
      consumption(resolution: $resolution, last: $last, before: $before) {
        pageInfo {
          startCursor
          hasPreviousPage
        }
        nodes {
          from
          to
          consumption
          cost
          unitPrice
          unitPriceVAT
          currency
        }
      }
    }
  }
}`

// QueryTimeZoneHome fetches the time zone of a specific home
const QueryTimeZoneHome = `query($homeId: ID!) {
  viewer {
    home(id: $homeId) {
      timeZone
    }
  }
}`

// QueryTimeZoneHomes fetches the time zone of all homes (used as fallback)
const QueryTimeZoneHomes = `{
  viewer {
    homes {
      timeZone
    }
  }
}`
