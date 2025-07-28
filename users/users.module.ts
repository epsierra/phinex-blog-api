import { Module } from '@nestjs/common';
import { DatabaseModule } from '../database/database.module';
import { UsersController } from './users.controller';
import { UsersService } from './users.service';
import { CommonModule } from '../common/common.module';
import { AuthModule } from '../auth/auth.module';
import { JwtModule } from '@nestjs/jwt';
import { oauthConfig } from '../config/oauth.config';
import { SmsModule } from 'src/sms/sms.module';
import { MailModule } from 'src/mail/mail.module';
import { NotificationsModule } from 'src/notifications/notifications.module';

@Module({
  imports: [
    DatabaseModule,
    CommonModule,
    AuthModule,
    SmsModule,
    MailModule,
    NotificationsModule,
    JwtModule.register({
      secret: oauthConfig.jwt.secret as string,
      signOptions: { expiresIn: oauthConfig.jwt.expiresIn },
    }),
  ],
  controllers: [UsersController],
  providers: [UsersService],
  exports: [UsersService],
})
export class UsersModule {}
