package internal

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// PostgreSQLClient represents the client to interact with PostgreSQL.
type PostgreSQLClient struct {
	DB *sql.DB
}

// ConnectPostgreSQL establishes a connection to PostgreSQL using the Database configuration.
func ConnectPostgreSQL(config AppConfig) (*PostgreSQLClient, error) {
	dbConfig := config.Database // Access the nested Database configuration directly
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", 
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	log.Println("Connected to PostgreSQL successfully")
	return &PostgreSQLClient{DB: db}, nil
}

// Close closes the database connection.
func (p *PostgreSQLClient) Close() error {
	return p.DB.Close()
}

// InsertDevice adds a new device to the database.
func (p *PostgreSQLClient) InsertDevice(device Device) error {
	query := `INSERT INTO devices (device_id, type, state) VALUES ($1, $2, $3)
              ON CONFLICT (device_id) DO UPDATE SET state = EXCLUDED.state`

	_, err := p.DB.Exec(query, device.ID, device.Type, device.State)
	if err != nil {
		return fmt.Errorf("failed to insert device: %w", err)
	}
	return nil
}

// UpdateDeviceState updates a device's state.
func (p *PostgreSQLClient) UpdateDeviceState(deviceID, state string) error {
	query := `UPDATE devices SET state = $1 WHERE device_id = $2`
	_, err := p.DB.Exec(query, state, deviceID)
	if err != nil {
		return fmt.Errorf("failed to update device state: %w", err)
	}
	return nil
}

// GetDevice retrieves a device's information.
func (p *PostgreSQLClient) GetDevice(deviceID string) (*Device, error) {
	query := `SELECT device_id, type, state FROM devices WHERE device_id = $1`
	row := p.DB.QueryRow(query, deviceID)

	var device Device
	err := row.Scan(&device.ID, &device.Type, &device.State)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No device found
		}
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	return &device, nil
}
