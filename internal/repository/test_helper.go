package repository

/*
type TestDB struct {
	Pool   *pgxpool.Pool
	Closer func()
}
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("orderflow_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to open sql.DB for migrations: %v", err)
	}

	migrationsDir, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatalf("failed to resolve migrations path: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}

	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		t.Fatalf("failed to run goose migrations: %v", err)
	}
	sqlDB.Close()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create pgxpool: %v", err)
	}

	closer := func() {
		pool.Close()
		if err := container.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}

	return &TestDB{
		Pool:   pool,
		Closer: closer,
	}
}
*/
