using Qwik.Model;
using Qwik.Model.Dto;

namespace Qwik;

public partial class VerifyEmailPage : ContentPage
{
    private readonly string Login;
    private readonly string Email;
	private readonly string Password;
    private readonly string SessionCode;

    public VerifyEmailPage(string login, string email, string password, string sessionCode)
    {
        this.Login = login;
        this.Email = email;
        this.Password = password;
        this.SessionCode = sessionCode;
        InitializeComponent();
    }

    private async void ConfirmRegClicked(object sender, EventArgs e)
    {
        try
        {
            string verifyCode = VerifyCode.Text;

            if (string.IsNullOrEmpty(verifyCode))
            {
                await DisplayAlert("Ошибка", "Введите код подтверждения", "ОК");
                return;
            }

            ConfirmRegReq req = new ConfirmRegReq
            {
                SessionCode = SessionCode,
                Login = Login,
                Password = Password,
                Email = Email,
                VerifyCode = verifyCode
            };

            ConfirmRegRes res = await Api.SendPostRequest<ConfirmRegReq, ConfirmRegRes>(req, Api.ConfirmRegEndpoint);

            if (string.IsNullOrEmpty(res.Id))
            {
                await DisplayAlert("Ошибка", "Неизветная ошибка", "ОК");
                return;
            }

            await Task.WhenAll(
                Shell.Current.Navigation.PopAsync(),
                Shell.Current.Navigation.PopAsync()
            );
        }
        catch (Exception ex)
        {
            await DisplayAlert("Ошибка", $"{ex}", "ОК");
            return;
        }
    }
}