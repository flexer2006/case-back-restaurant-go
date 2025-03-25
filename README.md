# Restaurant Booking System

API service for restaurant table booking with advanced notification and availability management features.

## Description

The system allows users to book tables at restaurants, and restaurants to manage their availability and interact with customers. Main features include:
- Restaurant data and schedule management
- Creation and management of bookings
- Email notification system
- Alternative booking time suggestions
- User account management
- Full API documentation via Swagger

## Technologies

- **Go 1.24.1**
- **PostgreSQL** - for data storage
- **Fiber v3** - high-performance web framework
- **Docker and Docker Compose** - for application containerization
- **Swagger/OpenAPI** - for API documentation
- **GoMail** - for sending email notifications
- **Zap** - for logging
- **Migrations** - for database structure management

## Installation and Launch

### Requirements

- Docker and Docker Compose
- Make (for convenient use of Makefile)

### Installation Steps

Clone the repository:
```bash
git clone https://github.com/flexer2006/case-back-restaurant-go.git
cd case-back-restaurant-go
```

Create .env file based on example.env:
```bash
cp example.env .env
```

Edit the .env file, setting necessary values for:
   - PostgreSQL connection
   - SMTP server settings for sending emails (host, port, username, password)

Launch the application using Make:
```bash
make run-all
```

Apply migrations to the database:
```bash
make migrate-up
```

### Checking Functionality

After launch, check server availability:
```bash
make health-check
```

## API Usage

### Swagger Documentation

Full API documentation is available at:
```
http://<host>:<port>/swagger-ui
```

Default: http://localhost:8080/swagger-ui

### Main Endpoints

#### Restaurants
- **GET /api/v1/restaurants** - Get list of restaurants
- **POST /api/v1/restaurants** - Create a restaurant
- **GET /api/v1/restaurants/{id}** - Get restaurant information
- **PUT /api/v1/restaurants/{id}** - Update restaurant information
- **DELETE /api/v1/restaurants/{id}** - Delete a restaurant

#### Schedule and Availability
- **POST /api/v1/restaurants/{id}/working-hours** - Set working hours
- **GET /api/v1/restaurants/{id}/working-hours** - Get working hours
- **POST /api/v1/restaurants/{id}/availability** - Set availability
- **GET /api/v1/restaurants/{id}/availability** - Get availability information

#### Bookings
- **POST /api/v1/bookings** - Create a booking
- **GET /api/v1/bookings/{id}** - Get booking information
- **POST /api/v1/bookings/{id}/confirm** - Confirm a booking
- **POST /api/v1/bookings/{id}/reject** - Reject a booking
- **POST /api/v1/bookings/{id}/cancel** - Cancel a booking
- **POST /api/v1/bookings/{id}/alternative** - Suggest alternative time

#### Users
- **POST /api/v1/users** - Create a user
- **GET /api/v1/users/{id}** - Get user information
- **PUT /api/v1/users/{id}** - Update user information
- **GET /api/v1/users/{id}/bookings** - Get user bookings
- **GET /api/v1/users/{id}/notifications** - Get user notifications

## Usage Examples

### Booking Lifecycle

**Create a booking**:
```
POST /api/v1/bookings
```

**Restaurant receives a notification** about a new booking via email.

**Booking confirmation** by the restaurant:
```
POST /api/v1/bookings/{id}/confirm
```

**User receives a notification** about booking confirmation.

### Alternative Suggestion

If the restaurant cannot confirm the booking for the requested time, it can **suggest an alternative**:
```
POST /api/v1/bookings/{id}/alternative
```

**User receives a notification** about the alternative time suggestion.

**User accepts the alternative suggestion**:
```
POST /api/v1/bookings/alternatives/{id}/accept
```

**Restaurant receives a notification** about accepting the alternative suggestion.

## Testing

To run tests:
```bash
make test
```

## Additional Commands

- `make down` - Stop all containers
- `make logs` - View logs
- `make clean` - Remove containers, images, and volumes
- `make help` - List of available commands
## Example .env
![image](https://github.com/user-attachments/assets/ac16e011-aea1-46a6-9316-f4afed77da9b)
