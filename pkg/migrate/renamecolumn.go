package migrate

import (
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

func tableExist(ctx context.Context, conn dialect.ExecQuerier, table string) (bool, error) {
	rows := sql.Rows{}

	query, args := sql.
		Select("datetime_precision").
		From(sql.Table("`information_schema`.`tables`")).
		Where(
			sql.And(
				sql.EQ("table_name", table),
			),
		).
		Count().
		Query()
	if err := conn.Query(ctx, query, args, &rows); err != nil {
		return false, err
	}

	for rows.Next() {
		count := 0
		if err := rows.Scan(&count); err != nil {
			return false, err
		}
		if count == 0 {
			rows.Close()
			return false, nil
		}
	}
	rows.Close()

	return true, nil
}

func columnExist(ctx context.Context, conn dialect.ExecQuerier, table, column string) (bool, error) {
	rows := sql.Rows{}

	query, args := sql.
		Select("datetime_precision").
		From(sql.Table("`information_schema`.`columns`")).
		Where(
			sql.And(
				sql.EQ("table_name", table),
				sql.EQ("column_name", column),
			),
		).
		Count().
		Query()

	if err := conn.Query(ctx, query, args, &rows); err != nil {
		return false, err
	}

	for rows.Next() {
		count := 0
		if err := rows.Scan(&count); err != nil {
			return false, err
		}
		if count > 0 {
			rows.Close()
			return true, nil
		}
	}
	rows.Close()

	return false, nil
}

func RenameColumn(ctx context.Context, conn dialect.ExecQuerier, table, srcColumn, dstColumn, columnType string, unsigned bool, nilable bool) error {
	if exist, err := tableExist(ctx, conn, table); err != nil || !exist {
		return err
	}
	if exist, err := columnExist(ctx, conn, table, srcColumn); err != nil || !exist {
		return err
	}
	if exist, err := columnExist(ctx, conn, table, dstColumn); err != nil || exist {
		return err
	}

	cb := sql.
		Column(dstColumn).
		Type(columnType)

	query, args := sql.
		AlterTable(table).
		ChangeColumn(srcColumn, cb).
		Query()
	if unsigned {
		query = query + " unsigned"
	}
	if !nilable {
		query = query + " NOT NULL"
	}
	if err := conn.Exec(ctx, query, args, nil); err != nil {
		return err
	}

	return nil
}
