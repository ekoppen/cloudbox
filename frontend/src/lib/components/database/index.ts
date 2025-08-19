// Database visualization components
export { default as SchemaVisualizer } from './schema-visualizer.svelte';
export { default as DataTable } from './data-table.svelte';
export { default as DatabaseNavigation } from './navigation.svelte';
export { default as RelationshipDiagram } from './relationship-diagram.svelte';

// Component types
export interface TableColumn {
  name: string;
  type: string;
  nullable: boolean;
  primary_key?: boolean;
  foreign_key?: {
    table: string;
    column: string;
  };
  default_value?: string;
  sortable?: boolean;
  width?: string;
}

export interface TableSchema {
  name: string;
  columns: TableColumn[];
  row_count: number;
  size: string;
  position?: { x: number; y: number };
  relationships?: {
    to_table: string;
    from_column: string;
    to_column: string;
    type: 'one-to-one' | 'one-to-many' | 'many-to-many';
  }[];
}

export interface TableData {
  columns: TableColumn[];
  rows: Record<string, any>[];
  total_count: number;
}

export interface DatabaseSection {
  id: string;
  name: string;
  icon: string;
  count?: number;
  badge?: string;
  active?: boolean;
  children?: DatabaseSection[];
}

export interface TableRelationship {
  id: string;
  from_table: string;
  from_column: string;
  to_table: string;
  to_column: string;
  constraint_name: string;
  relationship_type: 'one-to-one' | 'one-to-many' | 'many-to-many';
  on_delete: 'CASCADE' | 'RESTRICT' | 'SET NULL' | 'NO ACTION';
  on_update: 'CASCADE' | 'RESTRICT' | 'SET NULL' | 'NO ACTION';
}

export interface TableInfo {
  name: string;
  primary_keys: string[];
  foreign_keys: {
    column: string;
    references_table: string;
    references_column: string;
  }[];
  column_count: number;
  row_count: number;
}