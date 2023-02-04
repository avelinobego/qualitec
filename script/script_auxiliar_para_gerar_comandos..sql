-- -------------------------------------------------------------
-- Para gerar indices de time
-- -------------------------------------------------------------
select 
	CONCAT('ALTER TABLE `',
	tab.table_schema, '`.`',
	tab.table_name, 
	'` ADD INDEX `index_time` (`time`);') as command
from
	information_schema.tables as tab
inner join information_schema.columns as col
        on
	col.table_schema = tab.table_schema
	and col.table_name = tab.table_name
	and column_name = 'time'
where
	tab.table_type = 'BASE TABLE'
	and not EXISTS (
	select
		sta.*
	from
		information_schema.statistics sta
	where
		sta.table_schema not in ('information_schema', 'mysql', 'performance_schema', 'sys')
			and sta.INDEX_NAME = 'index_time'
			and sta.TABLE_NAME = tab.table_name
		group by
			sta.index_schema,
			sta.table_name
		order by
			sta.index_schema,
			sta.index_name)

order by
	tab.table_schema,
	tab.table_name;

-- -------------------------------------------------------------
-- Para gerar indices de time
-- -------------------------------------------------------------
select 
	CONCAT('ALTER TABLE `',
	tab.table_schema, '`.`',
	tab.table_name, 
	'` ADD INDEX `index_channel` (`channel`);') as command
from
	information_schema.tables as tab
inner join information_schema.columns as col
        on
	col.table_schema = tab.table_schema
	and col.table_name = tab.table_name
	and column_name = 'channel'
where
	tab.table_type = 'BASE TABLE'
	and not EXISTS (
	select
		sta.*
	from
		information_schema.statistics sta
	where
		sta.table_schema not in ('information_schema', 'mysql', 'performance_schema', 'sys')
			and sta.INDEX_NAME = 'index_channel'
			and sta.TABLE_NAME = tab.table_name
		group by
			sta.index_schema,
			sta.table_name
		order by
			sta.index_schema,
			sta.index_name)

order by
	tab.table_schema,
	tab.table_name;
