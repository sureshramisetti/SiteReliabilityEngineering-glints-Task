type baseBackupResult struct {
	StartPosition pglogrepl.LSN
	TimelineID    uint8
	Tablespaces   []baseBackupResultTablespace
}

type baseBackupResultTablespace struct {
	SpaceOID      *uint32
	SpaceLocation string
	Size          *uint8
}

func baseBackup(ctx context.Context, conn *pgconn.PgConn, output *io.Writer) error {
	var baseBackupResult baseBackupResult

	var buf []byte
	buf = (&pgproto3.Query{String: "BASE_BACKUP"}).Encode(buf)

	if err := conn.SendBytes(ctx, buf); err != nil {
		return fmt.Errorf("failed to send query: %v", err)
	}

	// The first ordinary result set contains the starting position of the
	// backup, in a single row with two columns. The first column contains the
	// start position given in XLogRecPtr format, and the second column
	// contains the corresponding timeline ID.
	result, err := readResult(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to read result: %w", err)
	}

	if len(result.Rows) != 1 {
		return fmt.Errorf("expected exactly 1 row in result set, %d found", len(result.Rows))
	}

	startPosition, err := pglogrepl.ParseLSN(string(result.Rows[0][0]))
	if err != nil {
		return fmt.Errorf("failed to parse start position: %w", err)
	}

	baseBackupResult.StartPosition = startPosition
	baseBackupResult.TimelineID = uint8(result.Rows[0][1][0])

	// The second ordinary result set has one row for each tablespace.
	result, err = readResult(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to read result: %w", err)
	}

	fieldIndices := make(map[string]int)
	for index, description := range result.FieldDescriptions {
		fieldIndices[string(description.Name)] = index
	}

	requiredFieldNames := []string{"spcoid", "spclocation", "size"}
	for _, name := range requiredFieldNames {
		if _, ok := fieldIndices[name]; !ok {
			return fmt.Errorf("expected field name \"%s\" not found in result set", name)
		}
	}

	for _, row := range result.Rows {
		var spcoid *uint32
		if len(row[fieldIndices["spcoid"]]) != 0 {
			*spcoid = binary.BigEndian.Uint32(row[fieldIndices["spcoid"]])
		}

		var spcSize *uint8
		if len(row[fieldIndices["size"]]) != 0 {
			*spcSize = uint8(row[fieldIndices["size"]][0])
		}

		baseBackupResult.Tablespaces = append(baseBackupResult.Tablespaces, baseBackupResultTablespace{
			SpaceOID:      spcoid,
			SpaceLocation: string(row[fieldIndices["spclocation"]]),
			Size:          spcSize,
		})
	}

	// Data dump follows.
	for {
		msg, err := conn.ReceiveMessage(ctx)
		if err != nil {
			return fmt.Errorf("failed to receive message: %w", err)
		}

		switch msg := msg.(type) {
		case *pgproto3.CopyOutResponse:
		case *pgproto3.CopyData:
			if _, err := (*output).Write(msg.Data); err != nil {
				return fmt.Errorf("failed to write to output: %w", err)
			}
		case *pgproto3.CopyDone:
			return nil
		default:
			return fmt.Errorf("Unexpected message encountered: %+v", msg)
		}

	}
}


func readResult(ctx context.Context, conn *pgconn.PgConn) (*pgconn.Result, error) {
	result := &pgconn.Result{}

readloop:
	for {
		msg, err := conn.ReceiveMessage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to receive message: %v", err)
		}

		switch msg := msg.(type) {
		case *pgproto3.RowDescription:
			result.FieldDescriptions = msg.Fields
		case *pgproto3.DataRow:
			result.Rows = append(result.Rows, msg.Values)
		case *pgproto3.ErrorResponse:
			return nil, pgconn.ErrorResponseToPgError(msg)
		case *pgproto3.CommandComplete:
			break readloop
		}
	}

	return result, nil
}
