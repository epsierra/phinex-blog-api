import { IsString, IsOptional, IsBoolean, IsEnum } from 'class-validator';
import { ApiProperty } from '@nestjs/swagger';
import { UserStatus, Role } from 'src/common/constants';

export class CreateUserDto {
  @ApiProperty({
    example: 'user@example.com',
    description: 'User email address',
  })
  @IsString()
  email: string;

  @ApiProperty({ example: 'Password123!', description: 'User password' })
  @IsString()
  password: string;

  @ApiProperty({
    example: 'John',
    description: 'User first name',
    required: false,
  })
  @IsString()
  @IsOptional()
  firstName?: string;

  @ApiProperty({
    example: 'Doe',
    description: 'User last name',
    required: false,
  })
  @IsString()
  @IsOptional()
  lastName?: string;

  @ApiProperty({
    example: 'John Doe',
    description: 'User full name',
    required: false,
  })
  @IsString()
  @IsOptional()
  fullName?: string;

  @ApiProperty({
    example: 'M',
    description: 'User middle name',
    required: false,
  })
  @IsString()
  @IsOptional()
  middleName?: string;

  @ApiProperty({
    example: '+1234567890',
    description: 'User phone number',
    required: false,
  })
  @IsString()
  @IsOptional()
  phoneNumber?: string;

  @ApiProperty({
    enum: UserStatus,
    example: UserStatus.Active,
    description: 'User status',
    required: false,
  })
  @IsEnum(UserStatus)
  @IsOptional()
  status?: UserStatus;

  @ApiProperty({
    example: true,
    description: 'Whether user is verified',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  verified?: boolean;

  @ApiProperty({
    example: true,
    description: 'Whether email is verified',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  emailIsVerified?: boolean;

  @ApiProperty({
    example: false,
    description: 'Whether phone number is verified',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  phoneNumberIsVerified?: boolean;

  @ApiProperty({ example: 'male', description: 'User gender', required: false })
  @IsString()
  @IsOptional()
  gender?: string;

  @ApiProperty({
    example: '1990-01-01',
    description: 'User date of birth',
    required: false,
  })
  @IsString()
  @IsOptional()
  dob?: string;

  @ApiProperty({
    enum: Role,
    example: 'Authenticated',
    description: 'Role to assign to the user',
  })
  @IsOptional()
  @IsEnum(Role)
  roleName?: Role;
}

export class UserUpdateDto {
  @ApiProperty({
    example: 'user@example.com',
    description: 'User email address',
    required: false,
  })
  @IsString()
  @IsOptional()
  email?: string;

  @ApiProperty({
    example: 'NewPassword123!',
    description: 'User password',
    required: false,
  })
  @IsString()
  @IsOptional()
  password?: string;

  @ApiProperty({
    example: 'John',
    description: 'User first name',
    required: false,
  })
  @IsString()
  @IsOptional()
  firstName?: string;

  @ApiProperty({
    example: 'Doe',
    description: 'User last name',
    required: false,
  })
  @IsString()
  @IsOptional()
  lastName?: string;

  @ApiProperty({
    example: 'John Doe',
    description: 'User full name',
    required: false,
  })
  @IsString()
  @IsOptional()
  fullName?: string;

  @ApiProperty({
    example: 'M',
    description: 'User middle name',
    required: false,
  })
  @IsString()
  @IsOptional()
  middleName?: string;

  @ApiProperty({
    example: '+1234567890',
    description: 'User phone number',
    required: false,
  })
  @IsString()
  @IsOptional()
  phoneNumber?: string;

  @ApiProperty({
    enum: UserStatus,
    example: UserStatus.Active,
    description: 'User status',
    required: false,
  })
  @IsEnum(UserStatus)
  @IsOptional()
  status?: UserStatus;

  @ApiProperty({
    example: true,
    description: 'Whether user is verified',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  verified?: boolean;

  @ApiProperty({
    example: true,
    description: 'Whether email is verified',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  emailIsVerified?: boolean;

  @ApiProperty({
    example: false,
    description: 'Whether phone number is verified',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  phoneNumberIsVerified?: boolean;

  @ApiProperty({ example: 'male', description: 'User gender', required: false })
  @IsString()
  @IsOptional()
  gender?: string;

  @ApiProperty({
    example: '1990-01-01',
    description: 'User date of birth',
    required: false,
  })
  @IsString()
  @IsOptional()
  dob?: string;
}

export class CreateUserSettingDto {
  @ApiProperty({
    example: true,
    description: 'Enable push notifications',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  pushNotificationsEnabled?: boolean;

  @ApiProperty({
    example: true,
    description: 'Enable email notifications',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  emailNotificationsEnabled?: boolean;

  @ApiProperty({
    example: 'light',
    description: 'User interface theme',
    required: false,
  })
  @IsString()
  @IsOptional()
  theme?: string;

  @ApiProperty({
    example: 'en',
    description: 'User language preference',
    required: false,
  })
  @IsString()
  @IsOptional()
  language?: string;

  @ApiProperty({
    example: 'public',
    description: 'Profile visibility setting',
    required: false,
  })
  @IsString()
  @IsOptional()
  profileVisibility?: string;

  @ApiProperty({
    example: false,
    description: 'Enable location sharing',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  locationSharingEnabled?: boolean;

  @ApiProperty({
    example: true,
    description: 'Enable auto refresh',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  autoRefreshEnabled?: boolean;

  @ApiProperty({
    example: {},
    description: 'Additional settings as JSON',
    required: false,
  })
  @IsOptional()
  settings?: any;
}

export class UpdateUserStatusDto {
  @ApiProperty({
    enum: UserStatus,
    example: UserStatus.Active,
    description: 'The new status for the user.',
  })
  @IsEnum(UserStatus)
  status: UserStatus;
}

export class SubscribeToPlanDto {
  @ApiProperty({
    enum: SubscriptionPlan,
    example: SubscriptionPlan.Monthly,
    description: 'The subscription plan to subscribe to.',
  })
  @IsEnum(SubscriptionPlan)
  plan: SubscriptionPlan;

  @ApiProperty({
    example: 'business_phi1234567890123456789',
    description: 'The ID of the business associated with the subscription (if applicable).',
    required: false,
  })
  @IsOptional()
  @IsString()
  businessId?: string;
}

export class UpdateUserSettingDto {
  @ApiProperty({
    example: true,
    description: 'Enable push notifications',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  pushNotificationsEnabled?: boolean;

  @ApiProperty({
    example: true,
    description: 'Enable email notifications',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  emailNotificationsEnabled?: boolean;

  @ApiProperty({
    example: 'dark',
    description: 'User interface theme',
    required: false,
  })
  @IsString()
  @IsOptional()
  theme?: string;

  @ApiProperty({
    example: 'es',
    description: 'User language preference',
    required: false,
  })
  @IsString()
  @IsOptional()
  language?: string;

  @ApiProperty({
    example: 'private',
    description: 'Profile visibility setting',
    required: false,
  })
  @IsString()
  @IsOptional()
  profileVisibility?: string;

  @ApiProperty({
    example: false,
    description: 'Enable location sharing',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  locationSharingEnabled?: boolean;

  @ApiProperty({
    example: true,
    description: 'Enable auto refresh',
    required: false,
  })
  @IsBoolean()
  @IsOptional()
  autoRefreshEnabled?: boolean;

  @ApiProperty({
    example: {},
    description: 'Additional settings as JSON',
    required: false,
  })
  @IsOptional()
  settings?: any;
}
