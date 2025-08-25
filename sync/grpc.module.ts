import { Module } from '@nestjs/common';
import { GrpcUserSyncService } from './grpc.service';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { join } from 'node:path';
import { DatabaseModule } from 'src/database/database.module';
import { ConfigModule } from 'src/config/config.module';
import { CommonModule } from 'src/common/common.module';

@Module({
  imports: [
    ClientsModule.register([
      {
        name: 'USER_PACKAGE',
        transport: Transport.GRPC,
        options: {
          package: 'user',
          url:"",
          protoPath: join(__dirname, 'sync.proto'),
          loader: {
            keepCase: true,
            longs: String,
            enums: String,
            defaults: true,
            oneofs: true,
          },
        },
      },
    ]),
    DatabaseModule,
    ConfigModule,
    CommonModule,
  ],
  providers: [GrpcUserSyncService],
  exports: [GrpcUserSyncService],
})
export class GrpcModule {}
