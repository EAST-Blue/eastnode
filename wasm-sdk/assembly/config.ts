import { JSON } from "assemblyscript-json/assembly";

export class Column {
  name: string;
  type: string;

  constructor(name: string, type: string) {
    this.name = name;
    this.type = type;
  }
}

export type TableSchema = Column[];

export function toSchema(tableDefinition: TableSchema): string {
  const obj = JSON.Value.Object();
  for (let i = 0; i < tableDefinition.length; i += 1) {
    obj.set(tableDefinition[i].name, tableDefinition[i].type);
  }
  return obj.toString();
}

export function toJson(jsonString: string): JSON.Obj {
  let jsonObj: JSON.Obj = <JSON.Obj>JSON.parse(jsonString);

  return jsonObj;
}
