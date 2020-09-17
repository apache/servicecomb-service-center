import {Component} from '@angular/core';
import {EditKieConfDialogComponent} from '../edit-kie-conf-dialog/edit-kie-conf-dialog.component';
import {MatDialog} from '@angular/material/dialog';

@Component({
  selector: 'app-settings-menu',
  templateUrl: './settings-menu.component.html',
  styleUrls: ['./settings-menu.component.scss']
})
export class SettingsMenuComponent {
  constructor(public dialog: MatDialog) {
  }

  openDialog(): void {
    const dialogRef = this.dialog.open(EditKieConfDialogComponent, {
      width: '550px'
    });

    dialogRef.afterClosed().subscribe(result => {
      console.log('The dialog was closed');
    });
  }
}
