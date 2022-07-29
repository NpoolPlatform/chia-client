package migrate

import (
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

func RenameColumn(ctx context.Context, conn dialect.ExecQuerier, table, srcColumn, dstColumn, columnType string, unsigned bool) error {
	query, args := sql.
		Select("datetime_precision").
		From(sql.Table("`information_schema`.`columns`")).
		Where(
			sql.And(
				sql.EQ("table_name", table),
				sql.EQ("column_name", dstColumn),
			),
		).
		Count().
		Query()

	rows := sql.Rows{}
	if err := conn.Query(ctx, query, args, &rows); err != nil {
		return err
	}

	for rows.Next() {
		count := 0
		if err := rows.Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			rows.Close()
			return nil
		}
	}
	rows.Close()

	cb := sql.
		Column(dstColumn).
		Type(columnType)

	query, args = sql.
		AlterTable(table).
		ChangeColumn(srcColumn, cb).
		Query()
	if unsigned {
		query = query + " unsigned"
	}
	if err := conn.Exec(ctx, query, args, nil); err != nil {
		return err
	}

	return nil
}
