select 
	CONCAT('ALTER TABLE `',
	tab.table_schema, '`.`',
	tab.table_name, 
	'` ADD INDEX IF NOT EXISTS `index_time` (`time`);') as command
from
	information_schema.tables as tab
inner join information_schema.columns as col
        on
	col.table_schema = tab.table_schema
	and col.table_name = tab.table_name
	and column_name = 'time'
where
	tab.table_type = 'BASE TABLE'
	and tab.TABLE_SCHEMA = 'mpm6861'
order by
	tab.table_schema,
	tab.table_name;
-- ------------------------------------------------------------------------------------------------
-- ------------------------------------------------------------------------------------------------
-- ------------------------------------------------------------------------------------------------
select 
	CONCAT('ALTER TABLE `',
	tab.table_schema, '`.`',
	tab.table_name, 
	'` ADD INDEX IF NOT EXISTS `index_channel` (`channel`);') as command
from
	information_schema.tables as tab
inner join information_schema.columns as col
        on
	col.table_schema = tab.table_schema
	and col.table_name = tab.table_name
	and column_name = 'channel'
where
	tab.table_type = 'BASE TABLE'
	and tab.TABLE_SCHEMA = 'earth1006'
order by
	tab.table_schema,
	tab.table_name;