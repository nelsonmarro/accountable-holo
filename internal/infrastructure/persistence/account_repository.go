func (r *AccountRepositoryImpl) GetPaginatedAccounts(ctx context.Context, page, pageSize int, filter string) (*domain.PaginatedResult[domain.Account], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}

	offset := (page - 1) * pageSize

	// Base query
	query := "from accounts where name ilike '%%' || $1 || '%%'"
	args := []interface{}{filter}

	// Get total count
	countQuery := "select count(*) " + query
	var totalCount int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total account count: %w", err)
	}

	// Get paginated accounts
	dataQuery := `select id, name, number, type, initial_balance, created_at, updated_at ` + query + ` order by name asc limit $2 offset $3`
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]domain.Account, 0, pageSize)
	for rows.Next() {
		var acc domain.Account
		err := rows.Scan(&acc.ID, &acc.Name, &acc.Number, &acc.Type, &acc.InitialBalance, &acc.CreatedAt, &acc.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, acc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over accounts: %w", err)
	}

	return &domain.PaginatedResult[domain.Account]{
		Data:       accounts,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}