import {Component, Input, OnChanges, OnInit, SimpleChanges, ViewChild} from '@angular/core';
import {MatTableDataSource} from '@angular/material/table';
import {SelectionModel} from '@angular/cdk/collections';
import {MatPaginator} from '@angular/material/paginator';
import {ServiceModel} from '../../../../models/service.model';
import {Observable} from 'rxjs';


@Component({
  selector: 'app-services-table',
  templateUrl: './services-table.component.html',
  styleUrls: ['./services-table.component.scss']
})
export class ServicesTableComponent implements OnInit, OnChanges {
  @Input() servicesTableData$: Observable<ServiceModel[]>;
  public displayedColumns: string[] = ['select', 'name', 'consumers', 'instances', 'relative', 'providers', 'kie'];
  public dataSource: MatTableDataSource<ServiceModel> = new MatTableDataSource<ServiceModel>([]);
  public selection = new SelectionModel<ServiceModel>(true, []);

  public isShowFilterInput = false;

  @ViewChild(MatPaginator, {static: true}) paginator: MatPaginator;

  public ngOnInit(): void {
    this.servicesTableData$.subscribe(d => {
      this.dataSource = new MatTableDataSource<ServiceModel>(d);
      this.dataSource.paginator = this.paginator;
    });

  }

  public ngOnChanges(changes: SimpleChanges): void {
  }

  /** Whether the number of selected elements matches the total number of rows. */
  public isAllSelected(): boolean {
    const numSelected = this.selection.selected.length;
    const numRows = this.dataSource.data.length;
    return numSelected === numRows;
  }

  /** Selects all rows if they are not all selected; otherwise clear selection. */
  public masterToggle(): void {
    this.isAllSelected() ?
      this.selection.clear() :
      this.dataSource.data.forEach(row => this.selection.select(row));
  }

  /** The label for the checkbox on the passed row */
  public checkboxLabel(row?: any): string {
    if (!row) {
      return `${this.isAllSelected() ? 'select' : 'deselect'} all`;
    }
    return `${this.selection.isSelected(row) ? 'deselect' : 'select'} row ${row.position + 1}`;
  }

  public applyFilter(event: Event): void {
    const filterValue = (event.target as HTMLInputElement).value;
    this.dataSource.filter = filterValue.trim().toLowerCase();
  }

  public showFilterInput(): void {
    this.isShowFilterInput = !this.isShowFilterInput;
    //this.dataSource = new MatTableDataSource<ServiceModel>(this.servicesTableData);
  }
}
