export class ServerEventModel {
  action: string;
  key: {tenant: string, appId: string, serviceName: string, version: string};
}
